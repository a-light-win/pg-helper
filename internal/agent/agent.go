package agent

import (
	"context"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/agent/grpc_handler"
	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/a-light-win/pg-helper/internal/job/db_job"
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Agent struct {
	Config *config.AgentConfig
	DbPool *pgxpool.Pool

	GrpcClient proto.DbTaskSvcClient

	JobProducer  *job.JobProducer
	JobScheduler *job.JobScheduler

	QuitCtx context.Context
	Quit    context.CancelFunc

	JobCtx  context.Context
	JobQuit context.CancelFunc
}

func New(config *config.AgentConfig) *Agent {
	agent := Agent{Config: config}
	return &agent
}

func (a *Agent) Init() error {
	a.QuitCtx, a.Quit = utils.InitSignalHandler()
	a.JobCtx, a.JobQuit = context.WithCancel(context.Background())

	if err := a.initJob(); err != nil {
		return err
	}
	if err := a.initDb(); err != nil {
		return err
	}
	if err := a.initGrpc(); err != nil {
		return err
	}

	grpc_handler.InitValidator()

	db_job.InitGlobalData(a.DbPool, a.JobCtx, &a.Config.Db)
	grpc_handler.InitGlobalData(a.DbPool, &a.Config.Db, a.GrpcClient, a.JobProducer, a.QuitCtx)

	return nil
}

func (a *Agent) Run() {
	a.runJob()

	a.runGrpc()

	a.JobScheduler.WaitGracefulShutdown()
}
