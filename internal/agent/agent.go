package agent

import (
	"context"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Agent struct {
	Config *config.AgentConfig
	DbPool *pgxpool.Pool

	GrpcClient     proto.DbTaskSvcClient
	GrpcCallCtx    context.Context
	CancelGrpcCall context.CancelFunc

	QuitCtx context.Context
	Quit    context.CancelFunc
}

func New(config *config.AgentConfig) *Agent {
	agent := Agent{Config: config}
	return &agent
}

func (a *Agent) Init() error {
	a.QuitCtx, a.Quit = utils.InitSignalHandler()
	if err := a.initDb(); err != nil {
		return err
	}
	if err := a.initGrpc(); err != nil {
		return err
	}
	return nil
}

func (a *Agent) Run() {
	<-a.QuitCtx.Done()
}
