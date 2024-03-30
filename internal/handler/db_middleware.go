package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

func (h *Handler) acquireDbConn(c *gin.Context) (*pgxpool.Conn, error) {
	conn, err := h.DbPool.Acquire(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire database connection")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service Unavailable: Failed to acquire database connection"})
	}
	return conn, err
}

func (h *Handler) DbConn(c *gin.Context) {
	conn, err := h.acquireDbConn(c)
	if err != nil {
		return
	}
	defer conn.Release()

	c.Set("DbConn", conn)
	c.Next()
}

func (h *Handler) DbSession(c *gin.Context) {
	conn, err := h.acquireDbConn(c)
	if err != nil {
		return
	}
	defer conn.Release()

	tx, err := conn.Begin(c)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start a transaction")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error: Failed to start a transaction"})
		return
	}
	defer tx.Rollback(c)

	c.Set("Tx", tx)
	c.Next()
}
