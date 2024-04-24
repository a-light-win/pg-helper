package server

import (
	"context"

	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/internal/handler/web_server"
	"github.com/a-light-win/pg-helper/internal/server/grpc_handler"
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type Server struct {
	Config *config.ServerConfig

	// web server
	Router  *gin.Engine
	Handler *web_server.Handler

	// grpc server
	GrpcServer *grpc.Server

	QuitCtx context.Context
}

func New(config *config.ServerConfig) *Server {
	r := gin.Default()

	server := Server{Config: config, Router: r}
	return &server
}

func (s *Server) Init() error {
	s.QuitCtx, _ = utils.InitSignalHandler()
	if err := s.initGrpc(); err != nil {
		return err
	}
	if err := s.initWebServer(); err != nil {
		return err
	}

	grpc_handler.InitGlobalData(&s.Config.Grpc, s.QuitCtx)

	return nil
}

func (s *Server) Run() {
	go s.runGrpcServer()

	go s.runWebServer()

	<-s.QuitCtx.Done()
	s.GrpcServer.GracefulStop()

	log.Log().Msg("Server is shutting down.")
}
