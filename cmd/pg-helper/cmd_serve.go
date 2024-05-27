package main

import (
	config "github.com/a-light-win/pg-helper/internal/config/server"
	server_ "github.com/a-light-win/pg-helper/internal/server"
	"github.com/a-light-win/pg-helper/pkg/utils/logger"
	"github.com/a-light-win/pg-helper/pkg/validate"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ServeCmd struct {
	config.ServerConfig
}

func (s *ServeCmd) Run() error {
	gin.SetMode(gin.ReleaseMode)

	validator := validate.New()
	if err := validator.Struct(&s.ServerConfig); err != nil {
		log.Error().Err(err).Msg("config validation failed")
		return err
	}

	logger.PrintCurrentLogLevel()
	log.Log().Msgf("pg-helper %s is start up", Version)

	server := server_.New(&s.ServerConfig)
	if err := server.Init(); err != nil {
		return err
	}

	server.Run()
	return nil
}
