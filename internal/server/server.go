package server

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/a-light-win/pg-helper/internal/handler"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
)

type Server struct {
	Config  *config.Config
	Router  *gin.Engine
	DbPool  *pgxpool.Pool
	Handler *handler.Handler

	JobProducer  *job.JobProducer
	JobScheduler *job.JobScheduler

	quit    context.CancelFunc
	QuitCtx context.Context
}

func New(config *config.Config) *Server {
	r := gin.Default()

	server := Server{Config: config, Router: r}
	return &server
}

func (s *Server) Init() error {
	s.initSignalHandler()
	err := s.initDb()
	if err != nil {
		return err
	}
	err = s.initJob()
	if err != nil {
		return err
	}
	return s.initWebServer()
}

func (s *Server) Run() {
	s.runJobScheduler()

	log.Log().Msg("Start the web server")

	// TODO: customize the address and port
	err := s.Router.Run() // listen and serve on 0.0.0.0:8080
	if err != nil {
		log.Error().Err(err).Msg("Web server exit with error")
	}

	s.WaitJobSchedulerExit()
}

func (s *Server) initSignalHandler() {
	s.QuitCtx, s.quit = context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case <-sigs:
			s.quit()
		}
	}()
}
