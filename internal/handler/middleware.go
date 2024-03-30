package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func (h *Handler) DbConn(c *gin.Context) {
	conn, err := h.DbPool.Acquire(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire database connection")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service Unavailable: Failed to acquire database connection"})
		return
	}
	defer conn.Release()

	c.Set("DbConn", conn)
	c.Next()
}

func (h *Handler) DbSession(c *gin.Context) {
	conn, err := h.DbPool.Acquire(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire database connection")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service Unavailable: Failed to acquire database connection"})
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to begin transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error: Failed to begin transaction"})
		return
	}
	defer tx.Rollback(c)

	log.Log().Msg("Set Tx to context")
	c.Set("Tx", tx)
	c.Next()
}
