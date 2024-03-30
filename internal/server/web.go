package server

import (
	"github.com/a-light-win/pg-helper/internal/handler"
	"github.com/a-light-win/pg-helper/internal/validate"
)

func (s *Server) initWebServer() error {
	s.Handler = handler.New(s.DbPool)

	validate.RegisterCustomValidations()

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
	s.Router.POST("/api/v1/db/create", s.Handler.DbConn, s.Handler.CreateDb)
	return nil
}
