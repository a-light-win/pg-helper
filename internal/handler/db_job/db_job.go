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

func (j *DbJob) UUID() uuid.UUID {
	return j.ID
}

func (j *DbJob) GetName() string {
	return j.ID.String()
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

func (j *DbJob) IsCancelling() bool {
	return j.Status == db.DbTaskStatusCancelling
}

func (j *DbJob) IsFailed() bool {
	switch j.Status {
	case db.DbTaskStatusFailed, db.DbTaskStatusCancelled:
		return true
	default:
		return false
	}
}

func (j *DbJob) Cancelling(reason string) {
	j.Status = db.DbTaskStatusCancelling
	j.Reason = reason
}
