package db_job

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/rs/zerolog/log"
)

func (j *DbJobHandler) RestoreDb(job *DbJob) error {
	log.Info().Str("DbName", job.DbName).
		Msg("Start to restore database")

	if _, err := os.Stat(filepath.Join(j.DbConfig.BackupRootPath, job.Data.BackupPath)); err != nil {
		log.Warn().Err(err).
			Str("DbName", job.DbName).
			Str("BackupPath", job.Data.BackupPath).
			Msg("Failed to restore database")
		return err
	}

	db_, err := j.DbApi.GetDb(job.DbID, nil)
	if err != nil {
		return err
	}

	if db_.Stage != proto.DbStage_CreateCompleted && db_.Stage != proto.DbStage_BackupCompleted {
		err := errors.New("db stage is not ready")
		log.Warn().Err(err).
			Str("DbName", job.DbName).
			Str("Stage", db_.Stage.String()).
			Str("Status", db_.Status.String()).
			Msg("Failed to restore database")
		return err
	}

	db_.Stage = proto.DbStage_Restoring
	db_.Status = proto.DbStatus_Processing
	j.DbApi.UpdateDbStatus(db_, nil)

	args := []string{
		"-h", j.DbConfig.Host(nil),
		"-p", fmt.Sprintf("%d", j.DbConfig.Port),
		"-U", j.DbConfig.User,
		"-d", job.DbName,
		"-f", job.Data.BackupPath,
	}

	cmd := exec.Command("psql", args...)
	cmd.Dir = j.DbConfig.BackupRootPath
	cmd.Stdin = strings.NewReader(j.DbConfig.Password() + "\n")
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		log.Error().Err(err).
			Strs("Args", args).
			Str("DbName", job.DbName).
			Str("BackupPath", job.Data.BackupPath).
			Str("StdErr", stdErr.String()).
			Msg("Failed to restore database")

		db_.Status = proto.DbStatus_Failed
		j.DbApi.UpdateDbStatus(db_, nil)

		job.Status = db.DbTaskStatusFailed
		job.Data.ErrReason = err.Error()
		j.DbApi.UpdateTaskStatus(job.DbTask, nil)
		return err
	}

	db_.Stage = proto.DbStage_RestoreCompleted
	db_.Status = proto.DbStatus_Done
	j.DbApi.UpdateDbStatus(db_, nil)

	job.Status = db.DbTaskStatusCompleted
	job.Data.ErrReason = ""
	j.DbApi.UpdateTaskStatus(job.DbTask, nil)
	return nil
}
