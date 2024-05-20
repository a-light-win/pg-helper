package server

import (
	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/internal/handler/grpc_server"
	"github.com/a-light-win/pg-helper/internal/handler/signal_server"
	"github.com/a-light-win/pg-helper/internal/handler/web_server"
	"github.com/a-light-win/pg-helper/pkg/handler"
)

type Server struct {
	handler.BaseServer
	Config *config.ServerConfig
}

func New(config *config.ServerConfig) *Server {
	ss := signal_server.NewSignalServer()
	gs := grpc_server.NewGrpcServer(&config.Grpc, ss.QuitCtx)
	ws := web_server.NewWebServer(&config.Web, gs.SvcHandler)

	server := Server{
		Config: config,
		BaseServer: handler.BaseServer{
			Name:    "Server",
			Servers: []handler.Server{ss, gs, ws},
			QuitCtx: ss.QuitCtx,
			Quit:    ss.Quit,
		},
	}
	return &server
}
