package db_task

import (
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/google/uuid"
)

type DbTask struct {
	*db.DbTask
	dbApi *db.DbApi

	*job.BaseTaskDependency
}

func NewDbTask(task *db.DbTask, dbApi *db.DbApi) job.Task {
	dbTask := &DbTask{DbTask: task, dbApi: dbApi}
	dbTask.BaseTaskDependency = job.NewBaseTaskDependency(dbTask, dbTask.Requires())
	return dbTask
}

func (t *DbTask) UUID() uuid.UUID {
	return t.ID
}

func (t *DbTask) JobID() uuid.UUID {
	return t.DbTask.JobID
}

func (t *DbTask) GetName() string {
	return t.ID.String()
}

func (t *DbTask) Requires() []uuid.UUID {
	return t.Data.DependsOn
}

func (t *DbTask) IsPending() bool {
	return t.Status == db.DbTaskStatusPending
}

func (t *DbTask) IsRunning() bool {
	return t.Status == db.DbTaskStatusRunning
}

func (t *DbTask) IsDone() bool {
	switch t.Status {
	case db.DbTaskStatusCompleted, db.DbTaskStatusFailed, db.DbTaskStatusCancelled:
		return true
	default:
		return false
	}
}

func (t *DbTask) IsFailed() bool {
	switch t.Status {
	case db.DbTaskStatusFailed, db.DbTaskStatusCancelled:
		return true
	default:
		return false
	}
}

func (t *DbTask) Cancel(reason string) {
	t.Status = db.DbTaskStatusCancelled
	t.Data.ErrReason = reason
	t.dbApi.UpdateTaskStatus(t.DbTask, nil)
}
