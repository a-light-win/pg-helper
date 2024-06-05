package web_server

import (
	"net/http"
	"time"

	"github.com/a-light-win/pg-helper/internal/interface/sourceApi"
	"github.com/gin-gonic/gin"
)

type CreateDbRequest struct {
	*sourceApi.DatabaseRequest
}

func NewCreateDbRequest() WebRequest {
	return &CreateDbRequest{}
}

func (r *CreateDbRequest) GetName() string {
	return "Create Database " + r.Name
}

func (r *CreateDbRequest) Scopes() []string {
	return []string{"db:write"}
}

func (r *CreateDbRequest) Resources() []string {
	return []string{"db:" + r.Name}
}

func (r *CreateDbRequest) AuthRequired() bool {
	return true
}

func (r *CreateDbRequest) Process(c *gin.Context, handler WebHandler) {
	h := handler.(*DbHandler)

	webSource := &sourceApi.DatabaseSource{
		DatabaseRequest: r.DatabaseRequest,
		Type:            sourceApi.WebSource,
	}
	webSource.State = sourceApi.SourceStateUnknown

	if err := h.SourceHandler.AddDatabaseSource(webSource); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ready := h.ReadyWaiter.WaitReady(webSource.InstanceName, webSource.Name, 5*time.Second)
	c.JSON(http.StatusOK, gin.H{"ready": ready})
}
