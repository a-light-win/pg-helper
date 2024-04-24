package server

import (
	"context"

	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/internal/handler"
	"github.com/a-light-win/pg-helper/internal/handler/grpc_server"
	"github.com/a-light-win/pg-helper/internal/handler/signal_server"
	"github.com/a-light-win/pg-helper/internal/handler/web_server"
	"github.com/rs/zerolog/log"
)

type Server struct {
	Config *config.ServerConfig

	Servers []handler.Server

	QuitCtx context.Context
}

func New(config *config.ServerConfig) *Server {
	ss := signal_server.NewSignalServer()
	gs := grpc_server.NewGrpcServer(&config.Grpc)
	ws := web_server.NewWebServer(&config.Web)

	server := Server{
		Config:  config,
		Servers: []handler.Server{ss, gs, ws},
		QuitCtx: ss.QuitCtx,
	}
	return &server
}

func (s *Server) Init() error {
	grpc_server.InitGlobalData(&s.Config.Grpc, s.QuitCtx)
	return nil
}

func (s *Server) Run() {
	s.run()

	<-s.QuitCtx.Done()

	s.Shutdown()
}

func (s *Server) run() {
	for _, server := range s.Servers {
		go server.Run()
	}

	log.Log().Msg("Server is running.")
}

func (s *Server) Shutdown() {
	waitExitCtx := context.Background()

	for i := len(s.Servers) - 1; i >= 0; i-- {
		s.Servers[i].Shutdown(waitExitCtx)
	}

	log.Log().Msg("Server is shutting down.")
}
