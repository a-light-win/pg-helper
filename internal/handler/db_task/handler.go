package db_task

import (
	"context"
	"errors"
	"fmt"

	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type DbTaskHandler struct {
	DbApi    *db.DbApi
	DbConfig *config.DbConfig

	jobProducer server.Producer
}

func NewDbTaskHandler(dbConfig *config.DbConfig) *DbTaskHandler {
	return &DbTaskHandler{
		DbConfig: dbConfig,
	}
}

func (h *DbTaskHandler) Handle(msg server.NamedElement) error {
	task, ok := msg.(*DbTask)
	if !ok {
		return errors.New("invalid job type")
	}

	err := h.handle(task)
	h.jobProducer.Send(task)
	return err
}

func (h *DbTaskHandler) handle(dbTask *DbTask) error {
	if dbTask.Status == db.DbTaskStatusCancelling {
		dbTask.Status = db.DbTaskStatusCancelled
		return h.DbApi.UpdateTaskStatus(dbTask.DbTask, nil)
	}

	switch dbTask.Action {
	case db.DbActionMigrateOut:
		return h.MigrateOut(dbTask)
	case db.DbActionCreateUser:
		return h.CreateUser(dbTask)
	case db.DbActionCreate:
		return h.CreateDatabase(dbTask)
	case db.DbActionBackup:
		return h.BackupDb(dbTask)
	case db.DbActionRestore:
		return h.RestoreDb(dbTask)
	case db.DbActionWaitReady:
		return h.WaitReadyDb(dbTask)
	default:
		return fmt.Errorf("invalid db action %s", dbTask.Action)
	}
}

func (h *DbTaskHandler) Init(setter server.GlobalSetter) (err error) {
	h.DbApi, err = db.NewDbApi(h.DbConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create db api")
		return err
	}
	setter.Set(constants.AgentKeyDbApi, h.DbApi)
	setter.Set(constants.AgentKeyConnCtx, h.DbApi.ConnCtx)

	return nil
}

func (h *DbTaskHandler) PostInit(getter server.GlobalGetter) error {
	h.DbApi.DbStatusNotifier = getter.Get(constants.AgentKeyNotifyDbStatusProducer).(server.Producer)
	h.jobProducer = getter.Get(constants.AgentKeyJobProducer).(server.Producer)
	quitCtx := getter.Get(constants.AgentKeyQuitCtx).(context.Context)

	if err := h.DbApi.MigrateDB(quitCtx); err != nil {
		return err
	}

	if err := h.recoverJobs(); err != nil {
		return err
	}

	return nil
}

func setFinalTaskStatus(api *db.DbApi, task *DbTask, err error) {
	if err != nil {
		task.Status = db.DbTaskStatusFailed
		task.Data.ErrReason = err.Error()
	} else {
		task.Status = db.DbTaskStatusCompleted
		task.Data.ErrReason = ""
	}
	api.UpdateTaskStatus(task.DbTask, nil)
}

func setFinalDbStatus(api *db.DbApi, db_ *db.Db, err error) {
	if err != nil {
		db_.Status = proto.DbStatus_Failed
		db_.ErrorMsg = err.Error()
	} else {
		db_.Status = proto.DbStatus_Done
		db_.ErrorMsg = ""
	}
	api.UpdateDbStatus(db_, nil)
}

func (h *DbTaskHandler) recoverJobs() error {
	log.Log().Msg("Recovering jobs ...")
	defer log.Log().Msg("Recovering jobs done")

	dbs, err := h.DbApi.ListDbs(nil)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	}

	for i := range dbs {
		if err := h.recoverJob(dbs[i].LastJobID, dbs[i].Name); err != nil {
			return err
		}
	}
	return nil
}

func (h *DbTaskHandler) recoverJob(jobID uuid.UUID, dbName string) error {
	if jobID == uuid.Nil {
		return nil
	}

	return h.DbApi.Query(func(q *db.Queries) error {
		tasks, err := q.ListDbTasksByJobID(h.DbApi.ConnCtx, jobID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil
			}
			return err
		}

		job_ := &job.BaseJob{ID: jobID}
		for i := range tasks {
			dbTask := NewDbTask(&tasks[i], h.DbApi)
			job_.Tasks = append(job_.Tasks, dbTask)
		}
		if job_.IsDone() {
			return nil
		}

		log.Info().
			Str("JobID", jobID.String()).
			Str("DbName", dbName).
			Msg("Recovering job")

		job_.Init()
		h.jobProducer.Send(job_)
		return nil
	})
}
