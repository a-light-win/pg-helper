package db_job

import (
	"errors"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/rs/zerolog/log"
)

func (j *DbJobHandler) WaitReadyDb(job *DbJob) (err error) {
	job.Status = db.DbTaskStatusRunning
	j.DbApi.UpdateTaskStatus(job.DbTask, nil)

	defer func() {
		if err != nil {
			job.Status = db.DbTaskStatusFailed
			job.Data.ErrReason = err.Error()
			j.DbApi.UpdateTaskStatus(job.DbTask, nil)
		}
	}()

	db_, err := j.DbApi.GetDb(job.DbID, nil)
	if err != nil {
		return err
	}

	if db_.Stage != proto.DbStage_CreateCompleted && db_.Stage != proto.DbStage_RestoreCompleted {
		err := errors.New("db stage is not ready")
		log.Warn().Err(err).
			Str("DbName", job.DbName).
			Str("Stage", db_.Stage.String()).
			Str("Status", db_.Status.String()).
			Msg("Failed to complete the db job")

		return err
	}

	db_.Stage = proto.DbStage_Ready
	db_.Status = proto.DbStatus_Done
	j.DbApi.UpdateDbStatus(db_, nil)

	job.Status = db.DbTaskStatusCompleted
	job.Data.ErrReason = ""
	j.DbApi.UpdateTaskStatus(job.DbTask, nil)
	return nil
}
