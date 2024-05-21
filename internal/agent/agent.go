package agent

import (
	"github.com/a-light-win/pg-helper/api/proto"
	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/internal/handler/db_job"
	"github.com/a-light-win/pg-helper/internal/handler/grpc_agent"
	"github.com/a-light-win/pg-helper/internal/handler/signal_server"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/a-light-win/pg-helper/pkg/handler"
)

type Agent struct {
	handler.BaseServer

	Config *config.AgentConfig
}

func New(config *config.AgentConfig) *Agent {
	signalServer := signal_server.NewSignalServer()

	dbStatusConsumer := handler.NewBaseConsumer[*proto.Database]("notify_db_status", &grpc_agent.DbStatusSender{}, 1)

	dbJobHandler := db_job.NewDbJobHandler(&config.Db)
	jobScheduler := job.NewJobScheduler(signalServer.QuitCtx, dbJobHandler, 8)

	grpcAgentServer := grpc_agent.NewGrpcAgentServer(&config.Grpc, signalServer.QuitCtx)

	agent := Agent{
		Config: config,
		BaseServer: handler.BaseServer{
			Name: "Agent",
			Servers: []handler.Server{
				signalServer,
				dbStatusConsumer,
				jobScheduler,
				grpcAgentServer,
			},
			QuitCtx: signalServer.QuitCtx,
			Quit:    signalServer.Quit,
		},
	}
	return &agent
}
