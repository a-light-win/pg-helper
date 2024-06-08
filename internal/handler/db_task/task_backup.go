package db_task

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/rs/zerolog/log"
)

func (h *DbTaskHandler) BackupDb(task *DbTask) (err error) {
	// Ensuere backup dir is exists
	os.MkdirAll(h.DbConfig.BackupDbDir(task.DbName), 0750)

	log.Info().Str("DbName", task.DbName).
		Msg("Start to backup database")

	task.Status = db.DbTaskStatusRunning
	h.DbApi.UpdateTaskStatus(task.DbTask, nil)
	defer func() { setFinalTaskStatus(h.DbApi, task, err) }()

	initial := task.Action != db.DbActionDailyBackup

	db_, err := h.DbApi.GetDb(task.DbID, nil)
	if err != nil {
		log.Error().Err(err).
			Str("DbName", task.DbName).
			Msg("Can not backup database due to db error")
		return err
	}

	if task.Action == db.DbActionDailyBackup && !db_.IsReadyToUse() {
		err := fmt.Errorf("db stage is not ready")
		log.Error().Err(err).
			Str("DbName", task.DbName).
			Str("Stage", db_.Stage.String()).
			Str("Status", db_.Status.String()).
			Msg("Can not backup the database")
		return err
	}

	if initial && !db_.ShouldBackup() {
		err := fmt.Errorf("can not trigger backup in current db stage")
		log.Error().Err(err).
			Str("DbName", task.DbName).
			Str("Stage", db_.Stage.String()).
			Str("Status", db_.Status.String()).
			Msg("Can not backup the database")
		return err
	}

	if initial {
		db_.Stage = proto.DbStage_BackupDatabase
		db_.Status = proto.DbStatus_Processing
		h.DbApi.UpdateDbStatus(db_, nil)
		defer func() { setFinalDbStatus(h.DbApi, db_, err) }()
	}

	// Backup the database here
	args := []string{
		"-h", h.DbConfig.Host(&config.InstanceInfo{InstanceName: task.Data.BackupFrom}),
		"-p", fmt.Sprint(h.DbConfig.Port),
		"-U", h.DbConfig.User,
		"-d", task.DbName,
		"-f", task.Data.BackupPath + ".tmp",
	}

	cmd := exec.Command("pg_dump", args...)
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
			Msg("Failed to backup database")

		os.Remove(filepath.Join(h.DbConfig.BackupRootPath, task.Data.BackupPath+".tmp"))
		return err
	}

	os.Rename(filepath.Join(h.DbConfig.BackupRootPath, task.Data.BackupPath+".tmp"),
		filepath.Join(h.DbConfig.BackupRootPath, task.Data.BackupPath))

	log.Log().Str("DbName", task.DbName).
		Str("BackupPath", task.Data.BackupPath).
		Msg("Database backup completed")

	return nil
}
