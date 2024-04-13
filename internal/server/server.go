package server

import (
	"context"

	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/a-light-win/pg-helper/internal/handler"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Server struct {
	Config  *config.Config
	Router  *gin.Engine
	DbPool  *pgxpool.Pool
	Handler *handler.Handler

	JobProducer  *job.JobProducer
	JobScheduler *job.JobScheduler

	QuitCtx context.Context
}

func New(config *config.Config) *Server {
	r := gin.Default()

	server := Server{Config: config, Router: r}
	return &server
}

func (s *Server) Init() error {
	s.QuitCtx, _ = utils.InitSignalHandler()
	return s.initWebServer()
}

func (s *Server) Run() {
	log.Log().Msg("Start the web server")

	// TODO: customize the address and port
	err := s.Router.Run() // listen and serve on 0.0.0.0:8080
	if err != nil {
		log.Error().Err(err).Msg("Web server exit with error")
	}
}
