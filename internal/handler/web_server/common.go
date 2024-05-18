package web_server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func logErrorAndRespond(c *gin.Context, err error, message string) {
	logErrorAndRespondWithCode(c, err, message, http.StatusInternalServerError)
}

func logErrorAndRespondWithCode(c *gin.Context, err error, message string, code int) {
	log.Error().
		Err(err).
		Msg(message)
	c.JSON(code, gin.H{"error": message})
}
