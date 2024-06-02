package web_server

import (
	"net/http"

	"github.com/a-light-win/pg-helper/internal/interface/grpc_server"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateDb(c *gin.Context) {
	var request grpc_server.CreateDbRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scopes := []string{"db:write"}
	resources := []string{"db:" + request.DbName}
	if err := h.authCheck(c, scopes, resources); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	db, err := h.DbManager.CreateDb(&request, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, db)
}

func (h *Handler) IsDbReady(c *gin.Context) {
	var request grpc_server.DbRequest
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scopes := []string{"db:read"}
	resources := []string{"db:" + request.DbName}
	if err := h.authCheck(c, scopes, resources); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	ready := h.DbManager.IsDbReady(&request)
	c.JSON(http.StatusOK, gin.H{"ready": ready})
}
