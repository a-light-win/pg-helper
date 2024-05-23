package server

import (
	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/internal/handler/grpc_server"
	"github.com/a-light-win/pg-helper/internal/handler/web_server"
	"github.com/a-light-win/pg-helper/pkg/server"
)

type Server struct {
	server.BaseServer
	Config *config.ServerConfig
}

func New(config *config.ServerConfig) *Server {
	ss := server.NewSignalServer()
	gs := grpc_server.NewGrpcServer(&config.Grpc, ss.QuitCtx)
	ws := web_server.NewWebServer(&config.Web, gs.SvcHandler)

	server := Server{
		Config: config,
		BaseServer: server.BaseServer{
			Name:    "Server",
			Servers: []server.Server{ss, gs, ws},
			QuitCtx: ss.QuitCtx,
			Quit:    ss.Quit,
		},
	}
	return &server
}
