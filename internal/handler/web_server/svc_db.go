package web_server

import (
	"net/http"

	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/gin-gonic/gin"
)

func (h *Handler) CreateDb(c *gin.Context) {
	var vo handler.CreateDbVO
	if err := c.ShouldBind(&vo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scopes := []string{"db:write"}
	resources := []string{"db:" + vo.DbName}
	if err := h.authCheck(c, scopes, resources); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	db, err := h.DbManager.CreateDb(&vo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, db)
}

func (h *Handler) IsDbReady(c *gin.Context) {
	var vo handler.DbVO
	if err := c.ShouldBind(&vo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	scopes := []string{"db:read"}
	resources := []string{"db:" + vo.DbName}
	if err := h.authCheck(c, scopes, resources); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	ready := h.DbManager.IsDbReady(&vo)
	c.JSON(http.StatusOK, gin.H{"ready": ready})
}