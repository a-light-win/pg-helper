package server

import (
	"context"

	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/internal/handler/grpc_server"
	"github.com/a-light-win/pg-helper/internal/handler/signal_server"
	"github.com/a-light-win/pg-helper/internal/handler/web_server"
	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/rs/zerolog/log"
)

type Server struct {
	Config *config.ServerConfig

	Servers []handler.Server

	QuitCtx context.Context
	Quit    context.CancelFunc

	globalVars map[string]interface{}
}

func New(config *config.ServerConfig) *Server {
	ss := signal_server.NewSignalServer()
	gs := grpc_server.NewGrpcServer(&config.Grpc, ss.QuitCtx)
	ws := web_server.NewWebServer(&config.Web, gs.SvcHandler)

	server := Server{
		Config:  config,
		Servers: []handler.Server{ss, gs, ws},
		QuitCtx: ss.QuitCtx,
		Quit:    ss.Quit,
	}
	return &server
}

func (s *Server) Init() error {
	for _, server := range s.Servers {
		if err := server.Init(s); err != nil {
			return err
		}
	}

	for _, server := range s.Servers {
		if err := server.PostInit(s); err != nil {
			return err
		}
	}

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

func (s *Server) Set(key string, value interface{}) {
	s.globalVars[key] = value
}

func (s *Server) Get(key string) interface{} {
	return s.globalVars[key]
}
