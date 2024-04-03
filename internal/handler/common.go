package handler

import (
	"net/http"

	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	DbPool      *pgxpool.Pool
	Config      *config.Config
	JobProducer *job.JobProducer
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
