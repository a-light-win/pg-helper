package server

import (
	"github.com/a-light-win/pg-helper/internal/handler/web_server"
	"github.com/a-light-win/pg-helper/internal/validate"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

func (s *Server) initWebServer() error {
	s.Handler = &web_server.Handler{
		Config: s.Config,
	}

	validatorEngine, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		log.Fatal().Msg("Failed to get validator engine")
	}
	validate.RegisterCustomValidations(validatorEngine)

	err := s.initWebServerByConfig()
	if err != nil {
		return err
	}
	return s.registerRoutes()
}

func (s *Server) initWebServerByConfig() error {
	s.Router.UseH2C = s.Config.Web.UseH2C
	s.Router.SetTrustedProxies(s.Config.Web.TrustedProxies)
	return nil
}

func (s *Server) registerRoutes() error {
	// dbGroup := s.Router.Group("/api/v1/db")

	// TODO: Get task status
	// dbGroup.GET("/migrate/:taskId", s.Handler.MigrateDbStatus)

	return nil
}

func (s *Server) runWebServer() {
	if !s.Config.Web.Enabled {
		log.Log().Msg("Web server is disabled")
		return
	}

	log.Log().Msg("Start the web server")
	// TODO: customize the address and port
	err := s.Router.Run() // listen and serve on 0.0.0.0:8080
	if err != nil {
		log.Error().Err(err).Msg("Web server exit with error")
	}
}
