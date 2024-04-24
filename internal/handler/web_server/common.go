package web_server

import (
	"net/http"

	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	DbPool *pgxpool.Pool
	Config *config.ServerConfig
}

func logErrorAndRespond(c *gin.Context, err error, message string) {
	logErrorAndRespondWithCode(c, err, message, http.StatusInternalServerError)
}

func logErrorAndRespondWithCode(c *gin.Context, err error, message string, code int) {
	log.Error().
		Err(err).
		Msg(message)
	c.JSON(code, gin.H{"error": message})
}
