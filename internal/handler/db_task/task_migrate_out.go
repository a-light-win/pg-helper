package db_task

import (
	"fmt"

	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/a-light-win/pg-helper/pkg/utils/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (h *DbTaskHandler) MigrateOut(task *DbTask) error {
	return h.DbApi.Query(func(q *db.Queries) error {
		return h.migrateOut(task, q)
	})
}

func (h *DbTaskHandler) migrateOut(task *DbTask, q *db.Queries) (err error) {
	log.Debug().Str("DbName", task.DbName).
		Msg("Start to migrate out database")

	task.Status = db.DbTaskStatusRunning
	h.DbApi.UpdateTaskStatus(task.DbTask, q)
	defer func() { setFinalTaskStatus(h.DbApi, task, err) }()

	db_, err := h.DbApi.GetDb(task.DbID, nil)
	if err != nil {
		log.Error().Err(err).
			Str("DbName", task.DbName).
			Msg("Can not restore database due to db error")
		return logger.NewAlreadyLoggedError(err, zerolog.ErrorLevel)
	}

	if !db_.IsReadyToUse() {
		err := fmt.Errorf("db stage is not ready")
		log.Error().Err(err).
			Str("DbName", task.DbName).
			Str("Stage", db_.Stage.String()).
			Str("Status", db_.Status.String()).
			Msg("Failed to migrate out database")
		return logger.NewAlreadyLoggedError(err, zerolog.ErrorLevel)
	}

	db_.Stage = proto.DbStage_Idle
	db_.Status = proto.DbStatus_Done
	h.DbApi.UpdateDbStatus(db_, q)
	return nil
}
