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

	doneJobProducer handler.Producer
}

func NewDbJobHandler(dbConfig *config.DbConfig) *DbJobHandler {
	return &DbJobHandler{
		DbConfig: dbConfig,
	}
}

func (h *DbJobHandler) Handle(msg handler.NamedElement) error {
	dbJob, ok := msg.(*DbJob)
	if !ok {
		return errors.New("invalid job type")
	}

	err := h.handle(dbJob)
	if err != nil {
		log.Error().Err(err).
			Str("JobName", dbJob.GetName()).
			Msg("Failed to handle job")
	}

	h.doneJobProducer.Send(dbJob)
	return err
}

func (h *DbJobHandler) handle(dbJob *DbJob) error {
	if dbJob.Status == db.DbTaskStatusCancelling {
		dbJob.Status = db.DbTaskStatusCancelled
		h.DbApi.UpdateTaskStatus(dbJob.UUID(), db.DbTaskStatusCancelled, dbJob.Reason)
		return nil
	}

	switch dbJob.Action {
	case db.DbActionBackup:
		return h.BackupDb(dbJob)
	case db.DbActionRestore:
		return h.RestoreDb(dbJob)
	case db.DbActionWaitReady:
		return h.WaitReadyDb(dbJob)
	default:
		return fmt.Errorf("invalid db action %s", dbJob.Action)
	}
}

func (h *DbJobHandler) RecoverJobs() (jobs []job.Job, err error) {
	err = h.DbApi.Query(func(q *db.Queries) error {
		activeTasks, err := q.ListActiveDbTasks(h.DbApi.ConnCtx)
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

func (h *DbJobHandler) Init(setter handler.GlobalSetter) (err error) {
	h.DbApi, err = db.NewDbApi(h.DbConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create db api")
		return err
	}
	setter.Set(constants.AgentKeyDbApi, h.DbApi)
	setter.Set(constants.AgentKeyConnCtx, h.DbApi.ConnCtx)

	return nil
}

func (h *DbJobHandler) PostInit(getter handler.GlobalGetter) error {
	h.DbApi.DbStatusNotifier = getter.Get(constants.AgentKeyNotifyDbStatusProducer).(handler.Producer)
	h.doneJobProducer = getter.Get(constants.AgentKeyDoneJobProducer).(handler.Producer)

	quitCtx := getter.Get(constants.AgentKeyQuitCtx).(context.Context)
	if err := h.DbApi.MigrateDB(quitCtx); err != nil {
		return err
	}

	return nil
}
