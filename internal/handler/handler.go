package handler

import (
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	Db db.DBTX
}

func (h *Handler) CreateDb(c *gin.Context) {
	// TODO: implement
}
