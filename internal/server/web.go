package server

import (
	"github.com/a-light-win/pg-helper/internal/handler"
	"github.com/a-light-win/pg-helper/internal/validate"
)

func (s *Server) initWebServer() error {
	s.Handler = &handler.Handler{
		DbPool:      s.DbPool,
		Config:      s.Config,
		JobProducer: s.JobProducer,
	}

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
	dbGroup := s.Router.Group("/api/v1/db")
	dbGroup.Use(s.Handler.DbConn)

	dbGroup.POST("create", s.Handler.CreateDb)

	dbGroup.POST("/backup", s.Handler.BackupDb)
	dbGroup.GET("/backup/:taskId", s.Handler.BackupStatus)

	dbGroup.POST("/migrate", s.Handler.MigrateDb)
	// TODO: Get task status
	// dbGroup.GET("/migrate/:taskId", s.Handler.MigrateDbStatus)

	return nil
}
