package web_server

import (
	"errors"
	"net/http"

	"github.com/a-light-win/pg-helper/internal/interface/grpcServerApi"
	"github.com/a-light-win/pg-helper/internal/interface/sourceApi"
	ginAuth "github.com/a-light-win/pg-helper/pkg/auth/gin"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type WebRequest interface {
	GetName() string

	Scopes() []string
	Resources() []string
	AuthRequired() bool

	Process(c *gin.Context, handler WebHandler)
}

type WebHandler interface {
	GetName() string
}

type (
	NewWebRequestFunc func() WebRequest
)

func WebHandleWrapper(handler WebHandler, newRequestFunc NewWebRequestFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		request := newRequestFunc()
		if err := c.ShouldBind(request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := authCheck(c, request); err != nil {
			c.JSON(403, gin.H{"error": err.Error()})
			return
		}
		request.Process(c, handler)
	}
}

func authCheck(c *gin.Context, request WebRequest) error {
	if !request.AuthRequired() {
		return nil
	}

	auth, ok := ginAuth.FetchAuthInfo(c)
	if !ok {
		return errors.New("no auth info")
	}

	scopes := request.Scopes()
	resources := request.Resources()

	if missing, ok := auth.ValidateScopes(scopes); !ok {
		log.Warn().Strs("missing", missing).
			Str("Name", request.GetName()).
			Strs("scopes", scopes).
			Strs("resources", resources).
			Msg("no scope permission")
		return errors.New("no scope permission")
	}

	if missing, ok := auth.ValidateResources(resources); !ok {
		log.Warn().Strs("missing", missing).
			Str("Name", request.GetName()).
			Strs("scopes", scopes).
			Strs("resources", resources).
			Msg("no resource permission")
		return errors.New("no resource permission")
	}
	return nil
}

type DbHandler struct {
	SourceHandler sourceApi.SourceHandler
	ReadyWaiter   grpcServerApi.DbReadyWaiter
}

func NewDbHandler(sourceHandler sourceApi.SourceHandler, readyWaiter grpcServerApi.DbReadyWaiter) *DbHandler {
	return &DbHandler{
		SourceHandler: sourceHandler,
		ReadyWaiter:   readyWaiter,
	}
}

func (h *DbHandler) GetName() string {
	return "Database Handler"
}
