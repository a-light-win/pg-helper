package db_task

import (
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/google/uuid"
)

type DbTask struct {
	*db.DbTask
}

func NewDbTask(task *db.DbTask) job.Job {
	return &DbTask{DbTask: task}
}

func (j *DbTask) UUID() uuid.UUID {
	return j.ID
}

func (j *DbTask) GetName() string {
	return j.ID.String()
}

func (j *DbTask) Requires() []uuid.UUID {
	return j.Data.DependsOn
}

func (j *DbTask) IsPending() bool {
	return j.Status == db.DbTaskStatusPending
}

func (j *DbTask) IsRunning() bool {
	return j.Status == db.DbTaskStatusRunning
}

func (j *DbTask) IsDone() bool {
	switch j.Status {
	case db.DbTaskStatusCompleted, db.DbTaskStatusFailed, db.DbTaskStatusCancelled:
		return true
	default:
		return false
	}
}

func (j *DbTask) IsCancelling() bool {
	return j.Status == db.DbTaskStatusCancelling
}

func (j *DbTask) IsFailed() bool {
	switch j.Status {
	case db.DbTaskStatusFailed, db.DbTaskStatusCancelled:
		return true
	default:
		return false
	}
}

func (j *DbTask) Cancelling(reason string) {
	j.Status = db.DbTaskStatusCancelling
	j.Reason = reason
}
