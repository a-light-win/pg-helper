package agent

import (
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
	ss := signal_server.NewSignalServer()

	dbJobHandler := db_job.NewDbJobHandler(&config.Db)
	js := job.NewJobScheduler(ss.QuitCtx, dbJobHandler, 8)

	gs := grpc_agent.NewGrpcAgentServer(&config.Grpc, ss.QuitCtx)

	agent := Agent{
		Config: config,
		BaseServer: handler.BaseServer{
			Name:    "Agent",
			Servers: []handler.Server{ss, js, gs},
			QuitCtx: ss.QuitCtx,
			Quit:    ss.Quit,
		},
	}
	return &agent
}
