package db_job

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/rs/zerolog/log"
)

type BackupDbJob struct {
	*DbJob
}

func (job *BackupDbJob) Run() {
	// Ensuere backup dir is exists
	os.MkdirAll(gd_.DbConfig.BackupDbDir(job.DbName), 0750)

	job.updateDbStatus(proto.DbStatus_Backuping)

	// Backup the database here
	args := []string{
		"-h", gd_.DbConfig.Host(gd_.DbConfig.CurrentVersion),
		"-p", fmt.Sprint(gd_.DbConfig.Port),
		"-U", gd_.DbConfig.User,
		"-d", job.DbName,
		"-f", job.Data.BackupPath + ".tmp",
	}

	cmd := exec.Command("pg_dump", args...)
	cmd.Dir = gd_.DbConfig.BackupRootPath
	cmd.Stdin = strings.NewReader(gd_.DbConfig.Password() + "\n")

	if err := cmd.Run(); err != nil {
		log.Error().Err(err).
			Str("DbName", job.DbName).
			Str("BackupPath", job.Data.BackupPath).
			Msg("Failed to backup database")

		os.Remove(filepath.Join(gd_.DbConfig.BackupRootPath, job.Data.BackupPath+".tmp"))

		job.updateTaskStatus(db.DbTaskStatusFailed, err.Error())
		return
	}

	os.Rename(filepath.Join(gd_.DbConfig.BackupRootPath, job.Data.BackupPath+".tmp"),
		filepath.Join(gd_.DbConfig.BackupRootPath, job.Data.BackupPath))

	log.Log().Str("DbName", job.DbName).
		Str("BackupPath", job.Data.BackupPath).
		Msg("Database backup completed")

	job.updateTaskStatus(db.DbTaskStatusCompleted, "")
}
