package db_job

import (
	"context"
	"errors"
	"fmt"

	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type DbJobHandler struct {
	DbApi    *db.DbApi
	DbConfig *config.DbConfig
}

func NewDbJobHandler(dbConfig *config.DbConfig) *DbJobHandler {
	return &DbJobHandler{
		DbConfig: dbConfig,
	}
}

func (j *DbJobHandler) Run(job job.Job) error {
	dbJob, ok := job.(*DbJob)
	if !ok {
		return errors.New("invalid job type")
	}

	switch dbJob.Action {
	case db.DbActionBackup:
		return j.BackupDb(dbJob)
	case db.DbActionRestore:
		return j.RestoreDb(dbJob)
	case db.DbActionWaitReady:
		return j.WaitReadyDb(dbJob)
	default:
		return fmt.Errorf("invalid db action %s", dbJob.Action)
	}
}

func (j *DbJobHandler) RecoverJobs() (jobs []job.Job, err error) {
	err = j.DbApi.Query(func(q *db.Queries) error {
		activeTasks, err := q.ListActiveDbTasks(j.DbApi.ConnCtx)
		if err != nil {
			if err != pgx.ErrNoRows {
				return err
			}
			return nil
		} else {
			jobs = make([]job.Job, 0, len(activeTasks))
			for _, task := range activeTasks {
				jobs = append(jobs, NewDbJob(&task))
			}
			return nil
		}
	})
	return
}

func (j *DbJobHandler) Cancel(job job.Job, reason string) error {
	dbJob, ok := job.(*DbJob)
	if !ok {
		return errors.New("invalid job type")
	}
	return j.DbApi.UpdateTaskStatus(dbJob.ID(), db.DbTaskStatusCancelled, reason)
}

func (j *DbJobHandler) Init(setter handler.GlobalSetter) (err error) {
	j.DbApi, err = db.NewDbApi(j.DbConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create db api")
		return err
	}
	setter.Set(constants.AgentKeyDbApi, j.DbApi)
	setter.Set(constants.AgentKeyConnCtx, j.DbApi.ConnCtx)

	return nil
}

func (j *DbJobHandler) PostInit(getter handler.GlobalGetter) error {
	j.DbApi.DbStatusNotifier = getter.Get(constants.AgentKeyNotifyDbStatusProducer).(handler.Producer)

	quitCtx := getter.Get(constants.AgentKeyQuitCtx).(context.Context)
	if err := j.DbApi.MigrateDB(quitCtx); err != nil {
		return err
	}

	return nil
}
