package agent

import (
	"context"

	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/internal/handler/db_job"
	"github.com/a-light-win/pg-helper/internal/handler/grpc_agent"
	"github.com/a-light-win/pg-helper/internal/handler/signal_server"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/rs/zerolog/log"
)

type Agent struct {
	Config *config.AgentConfig

	Servers []handler.Server

	QuitCtx context.Context
	Quit    context.CancelFunc

	globalVars map[string]interface{}
}

func New(config *config.AgentConfig) *Agent {
	ss := signal_server.NewSignalServer()

	dbJobHandler := db_job.NewDbJobHandler(&config.Db)
	js := job.NewJobScheduler(ss.QuitCtx, dbJobHandler, 8)

	gs := grpc_agent.NewGrpcAgentServer(&config.Grpc, ss.QuitCtx)

	agent := Agent{
		Config:  config,
		Servers: []handler.Server{ss, js, gs},
		QuitCtx: ss.QuitCtx,
		Quit:    ss.Quit,
	}
	return &agent
}

func (a *Agent) Init() error {
	for _, s := range a.Servers {
		if err := s.Init(a); err != nil {
			return err
		}
	}

	for _, s := range a.Servers {
		if err := s.PostInit(a); err != nil {
			return err
		}
	}

	return nil
}

func (a *Agent) Set(key string, value interface{}) {
	a.globalVars[key] = value
}

func (a *Agent) Get(key string) interface{} {
	return a.globalVars[key]
}

func (a *Agent) Run() {
	for _, server := range a.Servers {
		go server.Run()
	}

	log.Log().Msg("Agent is running")

	<-a.QuitCtx.Done()
	a.Quit()

	a.Shutdown()
}

func (a *Agent) Shutdown() {
	waitExitCtx := context.Background()

	for i := len(a.Servers) - 1; i >= 0; i-- {
		a.Servers[i].Shutdown(waitExitCtx)
	}

	log.Log().Msg("Agent is shutting down.")
}
