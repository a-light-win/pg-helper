package server

import (
	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/a-light-win/pg-helper/internal/handler"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Server struct {
	Config  *config.Config
	Router  *gin.Engine
	DbPool  *pgxpool.Pool
	Handler *handler.Handler
}

func New(config *config.Config) *Server {
	r := gin.Default()

	server := Server{Config: config, Router: r}
	return &server
}

func (s *Server) Init() error {
	err := s.initDb()
	if err != nil {
		return err
	}
	return s.initWebServer()
}

func (s *Server) Run() {
	log.Log().Msg("Start the web server")
	// TODO: Add graceful shutdown
	// TODO: customize the address and port
	err := s.Router.Run() // listen and serve on 0.0.0.0:8080
	if err != nil {
		log.Fatal().Err(err).Msg("Web server exit with error")
	}
}
