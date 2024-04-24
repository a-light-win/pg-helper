package server

import (
	"context"

	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/internal/handler"
	"github.com/a-light-win/pg-helper/internal/handler/grpc_server"
	"github.com/a-light-win/pg-helper/internal/handler/web_server"
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Server struct {
	Config *config.ServerConfig

	// web server
	WebServer handler.Server

	// grpc server
	GrpcServer *grpc.Server

	QuitCtx context.Context
}

func New(config *config.ServerConfig) *Server {
	server := Server{Config: config, WebServer: &web_server.WebServer{}}
	return &server
}

func (s *Server) Init() error {
	s.QuitCtx, _ = utils.InitSignalHandler()
	if err := s.initGrpc(); err != nil {
		return err
	}
	if err := s.WebServer.Init(&s.Config.Web); err != nil {
		return err
	}

	grpc_server.InitGlobalData(&s.Config.Grpc, s.QuitCtx)

	return nil
}

func (s *Server) Run() {
	go s.runGrpcServer()

	go s.WebServer.Run()

	<-s.QuitCtx.Done()

	s.WebServer.Shutdown(context.Background())
	s.GrpcServer.GracefulStop()

	log.Log().Msg("Server is shutting down.")
}
