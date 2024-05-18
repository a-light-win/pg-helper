package web_server

import (
	"errors"

	ginAuth "github.com/a-light-win/pg-helper/pkg/auth/gin"
	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	DbManager handler.DbManager
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) authCheck(c *gin.Context, scopes []string, resources []string) error {
	auth, ok := ginAuth.FetchAuthInfo(c)
	if !ok {
		return errors.New("no auth info")
	}
	if missing, ok := auth.ValidateScopes(scopes); !ok {
		log.Warn().Strs("missing", missing).
			Strs("scopes", scopes).
			Strs("resources", resources).
			Msg("no scope permission")
		return errors.New("no scope permission")
	}
	if missing, ok := auth.ValidateResources(resources); !ok {
		log.Warn().Strs("missing", missing).
			Strs("scopes", scopes).
			Strs("resources", resources).
			Msg("no resource permission")
		return errors.New("no resource permission")
	}
	return nil
}
