package db_job

import (
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/google/uuid"
)

type DbJob struct {
	*db.DbTask
}

func NewDbJob(task *db.DbTask) job.Job {
	return &DbJob{DbTask: task}
}

func (j *DbJob) ID() uuid.UUID {
	return j.DbTask.ID
}

func (j *DbJob) Name() string {
	return j.ID().String()
}

func (j *DbJob) Requires() []uuid.UUID {
	return j.Data.DependsOn
}

func (j *DbJob) IsPending() bool {
	return j.Status == db.DbTaskStatusPending
}

func (j *DbJob) IsRunning() bool {
	return j.Status == db.DbTaskStatusRunning
}

func (j *DbJob) IsDone() bool {
	switch j.Status {
	case db.DbTaskStatusCompleted, db.DbTaskStatusFailed, db.DbTaskStatusCancelled:
		return true
	default:
		return false
	}
}

func (j *DbJob) IsFailed() bool {
	switch j.Status {
	case db.DbTaskStatusFailed, db.DbTaskStatusCancelled:
		return true
	default:
		return false
	}
}
