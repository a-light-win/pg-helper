package web_server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IsDbReadyRequest struct {
	Name         string `form:"name" json:"name" binding:"max=63,id"`
	InstanceName string `form:"instance_name" json:"instance_name" binding:"max=63,iname"`
}

func NewIsDbReadyRequest() WebRequest {
	return &IsDbReadyRequest{}
}

func (r *IsDbReadyRequest) GetName() string {
	return fmt.Sprintf("Is Db (%s) Ready", r.Name)
}

func (r *IsDbReadyRequest) Scopes() []string {
	return []string{"db:read"}
}

func (r *IsDbReadyRequest) Resources() []string {
	return []string{"db:" + r.Name}
}

func (r *IsDbReadyRequest) AuthRequired() bool {
	return true
}

func (r *IsDbReadyRequest) Process(c *gin.Context, handler WebHandler) {
	h := handler.(*DbHandler)

	ready := h.SourceHandler.IsReady(r.Name, r.InstanceName)
	c.JSON(http.StatusOK, gin.H{"ready": ready})
}
