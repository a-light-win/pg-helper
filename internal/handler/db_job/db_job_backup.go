package db_job

import (
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

func (j *DbJobHandler) BackupDb(job *DbJob) error {
	// Ensuere backup dir is exists
	os.MkdirAll(j.DbConfig.BackupDbDir(job.DbName), 0750)

	db_, err := j.DbApi.GetDb(job.DbID, nil)
	if err != nil {
		return err
	}

	var initial bool
	if db_.Stage == proto.DbStage_Creating ||
		db_.Stage == proto.DbStage_CreateCompleted ||
		db_.Stage == proto.DbStage_BackingUp {
		initial = true
	}

	if initial {
		db_.Stage = proto.DbStage_BackingUp
		db_.Status = proto.DbStatus_Processing
		if err := j.DbApi.UpdateDbStatus(db_, nil); err != nil {
			log.Warn().Err(err).
				Str("DbName", db_.Name).
				Msg("Failed to update db status")
		}
	}

	// Backup the database here
	args := []string{
		"-h", j.DbConfig.Host(&config.InstanceInfo{InstanceName: job.Data.BackupFrom}),
		"-p", fmt.Sprint(j.DbConfig.Port),
		"-U", j.DbConfig.User,
		"-d", job.DbName,
		"-f", job.Data.BackupPath + ".tmp",
	}

	cmd := exec.Command("pg_dump", args...)
	cmd.Dir = j.DbConfig.BackupRootPath
	cmd.Stdin = strings.NewReader(j.DbConfig.Password + "\n")

	if err := cmd.Run(); err != nil {
		log.Error().Err(err).
			Strs("Args", args).
			Str("DbName", job.DbName).
			Str("BackupPath", job.Data.BackupPath).
			Msg("Failed to backup database")

		os.Remove(filepath.Join(j.DbConfig.BackupRootPath, job.Data.BackupPath+".tmp"))

		db_.Status = proto.DbStatus_Failed
		j.DbApi.UpdateDbStatus(db_, nil)

		job.Status = db.DbTaskStatusFailed
		job.Data.ErrReason = err.Error()
		j.DbApi.UpdateTaskStatus(job.DbTask, nil)
		return nil
	}

	os.Rename(filepath.Join(j.DbConfig.BackupRootPath, job.Data.BackupPath+".tmp"),
		filepath.Join(j.DbConfig.BackupRootPath, job.Data.BackupPath))

	log.Log().Str("DbName", job.DbName).
		Str("BackupPath", job.Data.BackupPath).
		Msg("Database backup completed")

	db_.Stage = proto.DbStage_BackupCompleted
	db_.Status = proto.DbStatus_Done
	j.DbApi.UpdateDbStatus(db_, nil)

	job.Status = db.DbTaskStatusCompleted
	job.Data.ErrReason = ""
	j.DbApi.UpdateTaskStatus(job.DbTask, nil)
	return nil
}
