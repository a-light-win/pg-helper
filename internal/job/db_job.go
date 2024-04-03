package job

import (
	"context"

	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DbJob struct {
	DbTask *db.DbTask

	// ReadyChan will trigger when the job is done,
	// The consumer can wait on this channel to know when the job is done.
	ReadyChan chan struct{}

	DbPool *pgxpool.Pool
	ctx    context.Context
}

func NewDbJob(ctx context.Context, task *db.DbTask, pool *pgxpool.Pool) *DbJob {
	return &DbJob{
		DbTask:    task,
		ReadyChan: make(chan struct{}),
		DbPool:    pool,
		ctx:       ctx,
	}
}

func (j *DbJob) ID() uuid.UUID {
	return j.DbTask.ID
}

func (j *DbJob) Name() string {
	return j.ID().String()
}

func (j *DbJob) Requires() []uuid.UUID {
	return j.DbTask.Data.DependsOn
}

func (j *DbJob) Run() {
	defer close(j.ReadyChan)

	switch j.DbTask.Action {
	// TODO: add db.DbActionBackup
	}

	j.ReadyChan <- struct{}{}
}

func (j *DbJob) Cancel(reason string) error {
	return j.updateStatus(db.DbTaskStatusCancelled, reason)
}

func (j *DbJob) updateStatus(status db.DbTaskStatus, reason string) error {
	conn, err := j.DbPool.Acquire(j.ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	q := db.New(conn)
	if err = q.SetDbTaskStatus(j.ctx, db.SetDbTaskStatusParams{ID: j.ID(), Status: status, ErrReason: reason}); err != nil {
		return err
	}

	return nil
}

func (j *DbJob) IsPending() bool {
	return j.DbTask.Status == db.DbTaskStatusPending
}

func (j *DbJob) IsRunning() bool {
	return j.DbTask.Status == db.DbTaskStatusRunning
}

func (j *DbJob) IsDone() bool {
	switch j.DbTask.Status {
	case db.DbTaskStatusCompleted, db.DbTaskStatusFailed, db.DbTaskStatusCancelled:
		return true
	default:
		return false
	}
}

func (j *DbJob) IsFailed() bool {
	switch j.DbTask.Status {
	case db.DbTaskStatusFailed, db.DbTaskStatusCancelled:
		return true
	default:
		return false
	}
}
