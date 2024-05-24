package server

import (
	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/internal/handler/grpc_server"
	"github.com/a-light-win/pg-helper/internal/handler/web_server"
	"github.com/a-light-win/pg-helper/internal/source"
	"github.com/a-light-win/pg-helper/pkg/server"
)

type Server struct {
	server.BaseServer
	Config *config.ServerConfig
}

func New(config *config.ServerConfig) *Server {
	signalServer := server.NewSignalServer()
	grpcServer := grpc_server.NewGrpcServer(&config.Grpc, signalServer.QuitCtx)
	webServer := web_server.NewWebServer(&config.Web, grpcServer.SvcHandler)
	cronServer := server.NewCronServer()

	sourceHandler := source.NewSourceHandler(&config.Source)
	sourceConsumer := server.NewBaseConsumer[server.NamedElement]("Source Manager", sourceHandler, 8)

	fileSourceHandler := source.NewFileSourceHandler(sourceHandler)
	fileSourceMonitor := server.NewFileMonitor("File Source Monitor", fileSourceHandler)

	pgServer := Server{
		Config: config,
		BaseServer: server.BaseServer{
			Name: "PG Helper Server",
			Servers: []server.Server{
				signalServer,
				cronServer,
				grpcServer,
				sourceConsumer,
				fileSourceMonitor,
				webServer,
			},
			QuitCtx: signalServer.QuitCtx,
			Quit:    signalServer.Quit,
		},
	}

	pgServer.Set(constants.ServerKeyCronProducer, cronServer.Producer())
	pgServer.Set(constants.ServerKeySourceProducer, sourceConsumer.Producer())

	return &pgServer
}
