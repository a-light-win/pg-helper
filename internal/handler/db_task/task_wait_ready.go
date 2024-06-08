package db_task

import (
	"errors"

	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/rs/zerolog/log"
)

func (h *DbTaskHandler) WaitReadyDb(task *DbTask) (err error) {
	task.Status = db.DbTaskStatusRunning
	h.DbApi.UpdateTaskStatus(task.DbTask, nil)

	defer func() { setFinalTaskStatus(h.DbApi, task, err) }()

	db_, err := h.DbApi.GetDb(task.DbID, nil)
	if err != nil {
		log.Warn().Err(err).
			Str("DbName", task.DbName).
			Msg("Can not complete the task due to db error")
		return err
	}

	if db_.Status == proto.DbStatus_Done &&
		(db_.Stage == proto.DbStage_CreateDatabase ||
			db_.Stage == proto.DbStage_RestoreDatabase) {
		db_.Stage = proto.DbStage_ReadyToUse
		h.DbApi.UpdateDbStatus(db_, nil)
		return nil
	}

	err = errors.New("db stage is not ready")
	log.Warn().Err(err).
		Str("DbName", task.DbName).
		Str("Stage", db_.Stage.String()).
		Str("Status", db_.Status.String()).
		Msg("Failed to complete the db job")

	return err
}
