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
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
)

type DbTaskHandler struct {
	DbApi    *db.DbApi
	DbConfig *config.DbConfig

	doneJobProducer server.Producer
}

func NewDbTaskHandler(dbConfig *config.DbConfig) *DbTaskHandler {
	return &DbTaskHandler{
		DbConfig: dbConfig,
	}
}

func (h *DbTaskHandler) Handle(msg server.NamedElement) error {
	dbJob, ok := msg.(*DbTask)
	if !ok {
		return errors.New("invalid job type")
	}

	err := h.handle(dbJob)
	h.doneJobProducer.Send(dbJob)
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

func (h *DbTaskHandler) RecoverJobs() (jobs []job.Job, err error) {
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
				jobs = append(jobs, NewDbTask(&task))
			}
			return nil
		}
	})
	return
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
	h.doneJobProducer = getter.Get(constants.AgentKeyDoneJobProducer).(server.Producer)

	quitCtx := getter.Get(constants.AgentKeyQuitCtx).(context.Context)
	if err := h.DbApi.MigrateDB(quitCtx); err != nil {
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
