package db_task

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/rs/zerolog/log"
)

func (h *DbTaskHandler) RestoreDb(task *DbTask) (err error) {
	log.Info().Str("DbName", task.DbName).
		Msg("Start to restore database")

	task.Status = db.DbTaskStatusRunning
	h.DbApi.UpdateTaskStatus(task.DbTask, nil)
	defer func() { setFinalTaskStatus(h.DbApi, task, err) }()

	db_, err := h.DbApi.GetDb(task.DbID, nil)
	if err != nil {
		log.Error().Err(err).
			Str("DbName", task.DbName).
			Msg("Can not restore database due to db error")
		return err
	}

	if !db_.CanRestore() {
		err := errors.New("db stage is not ready")
		log.Error().Err(err).
			Str("DbName", task.DbName).
			Str("Stage", db_.Stage.String()).
			Str("Status", db_.Status.String()).
			Msg("Failed to restore database")
		return err
	}

	db_.Stage = proto.DbStage_RestoreDatabase
	db_.Status = proto.DbStatus_Processing
	h.DbApi.UpdateDbStatus(db_, nil)

	defer func() { setFinalDbStatus(h.DbApi, db_, err) }()

	if _, err := os.Stat(filepath.Join(h.DbConfig.BackupRootPath, task.Data.BackupPath)); err != nil {
		log.Warn().Err(err).
			Str("DbName", task.DbName).
			Str("BackupPath", task.Data.BackupPath).
			Msg("Failed to restore database")
		return err
	}

	args := []string{
		"-h", h.DbConfig.Host(nil),
		"-p", fmt.Sprintf("%d", h.DbConfig.Port),
		"-U", h.DbConfig.User,
		"-d", task.DbName,
		"-f", task.Data.BackupPath,
	}

	cmd := exec.Command("psql", args...)
	cmd.Dir = h.DbConfig.BackupRootPath
	cmd.Stdin = strings.NewReader(h.DbConfig.Password + "\n")
	var stdErr bytes.Buffer
	cmd.Stderr = &stdErr

	if err := cmd.Run(); err != nil {
		log.Error().Err(err).
			Strs("Args", args).
			Str("DbName", task.DbName).
			Str("BackupPath", task.Data.BackupPath).
			Str("StdErr", stdErr.String()).
			Msg("Failed to restore database")
		return err
	}

	return nil
}
