package agent

import (
	"github.com/a-light-win/pg-helper/api/proto"
	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/internal/constants"
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

	dbStatusConsumer := handler.NewBaseConsumer[*proto.Database]("Db Status Notifier", &grpc_agent.DbStatusSender{}, 1)

	dbJobHandler := db_job.NewDbJobHandler(&config.Db)
	dbJobConsumer := handler.NewBaseConsumer[job.Job]("Db Job Handler", dbJobHandler, 4)

	pendingJobHandler := &job.PendingJobHandler{}
	pendingJobConsumer := handler.NewBaseConsumer[job.Job]("Pending Job Handler", pendingJobHandler, 1)
	doneJobConsumer := handler.NewBaseConsumer[job.Job]("Done Job Handler", &job.DoneJobHandler{pendingJobHandler}, 1)

	grpcAgentServer := grpc_agent.NewGrpcAgentServer(&config.Grpc, signalServer.QuitCtx)

	agent := Agent{
		Config: config,
		BaseServer: handler.BaseServer{
			Name: "Agent",
			Servers: []handler.Server{
				signalServer,
				dbStatusConsumer,
				dbJobConsumer,
				doneJobConsumer,
				pendingJobConsumer,
				grpcAgentServer,
			},
			QuitCtx: signalServer.QuitCtx,
			Quit:    signalServer.Quit,
		},
	}

	agent.Set(constants.AgentKeyNotifyDbStatusProducer, dbStatusConsumer.Producer())
	agent.Set(constants.AgentKeyJobProducer, pendingJobConsumer.Producer())
	agent.Set(constants.AgentKeyDoneJobProducer, doneJobConsumer.Producer())
	agent.Set(constants.AgentKeyReadyToRunJobProducer, dbJobConsumer.Producer())

	return &agent
}
