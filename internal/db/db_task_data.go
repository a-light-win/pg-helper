package db

import "github.com/google/uuid"

type DbTaskData struct {
	// This task depends on other tasks
	// Valid in all tasks
	DependsOn []uuid.UUID `json:"depends_on"`

	// When the task is failed or canceled,
	// this field contains the reason of the failure
	// Valid in all tasks
	ErrReason string `json:"err_reason"`

	// The pg instance to run backup task
	BackupFrom string `json:"backup_from"`
	// The backup path of the database, in format of `pg-<major>/<database name>/<timestamp>.sql`
	// e.g. `pg-12/mydb/2021-01-01T01:01:01.sql`
	//
	// This path is relative to DbConfig.BackupRootPath
	//
	// Valid in following tasks:
	// - backup
	// - remote-backup
	// - restore
	BackupPath string `json:"backup_path"`
}
