package agent

import (
	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/internal/handler/db_task"
	"github.com/a-light-win/pg-helper/internal/handler/grpc_agent"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/a-light-win/pg-helper/pkg/server"
)

type Agent struct {
	server.BaseServer

	Config *config.AgentConfig
}

func New(config *config.AgentConfig) *Agent {
	signalServer := server.NewSignalServer()

	dbStatusConsumer := server.NewBaseConsumer[*proto.Database]("Db Status Notifier", &grpc_agent.DbStatusSender{}, 1)

	dbJobHandler := db_task.NewDbTaskHandler(&config.Db)
	dbJobConsumer := server.NewBaseConsumer[job.Task]("Db Job Handler", dbJobHandler, 4)

	jobHandler := &job.JobHandler{}
	jobConsumer := server.NewBaseConsumer[server.NamedElement]("Pending Job Handler", jobHandler, 1)

	grpcAgentServer := grpc_agent.NewGrpcAgentServer(&config.Grpc, signalServer.QuitCtx)

	agent := Agent{
		Config: config,
		BaseServer: server.BaseServer{
			Name: "PG Helper Agent",
			Servers: []server.Server{
				signalServer,
				dbStatusConsumer,
				dbJobConsumer,
				jobConsumer,
				grpcAgentServer,
			},
			QuitCtx: signalServer.QuitCtx,
			Quit:    signalServer.Quit,
		},
	}

	agent.Set(constants.AgentKeyNotifyDbStatusProducer, dbStatusConsumer.Producer())
	agent.Set(constants.AgentKeyReadyToRunJobProducer, dbJobConsumer.Producer())
	agent.Set(constants.AgentKeyJobProducer, jobConsumer.Producer())

	return &agent
}
