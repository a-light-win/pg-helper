package db_job

import (
	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type DbJob struct {
	*db.DbTask

	log *zerolog.Logger
}

func (j *DbJob) Log() *zerolog.Logger {
	if j.log == nil {
		logger := log.With().
			Str("TaskId", j.ID().String()).
			Str("DbName", j.DbName).
			Str("Action", string(j.Action)).
			Logger()
		j.log = &logger
	}
	return j.log
}

func NewDbJob(task *db.DbTask) job.Job {
	dbJob := &DbJob{DbTask: task}

	switch task.Action {
	case db.DbActionBackup:
		return &BackupDbJob{DbJob: dbJob}
	case db.DbActionRestore:
		return &RestoreDbJob{DbJob: dbJob}
	case db.DbActionWaitReady:
		return &WaitReadyDbJob{DbJob: dbJob}
	}
	return nil
}

func RecoverJobs() ([]job.Job, error) {
	conn, err := gd_.DbPool.Acquire(gd_.ConnCtx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	q := db.New(conn)
	activeTasks, err := q.ListActiveDbTasks(gd_.ConnCtx)
	if err != nil {
		if err != pgx.ErrNoRows {
			return nil, err
		}
		return nil, nil
	} else {
		jobs := make([]job.Job, 0, len(activeTasks))
		for _, task := range activeTasks {
			jobs = append(jobs, NewDbJob(&task))
		}
		return jobs, nil
	}
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

func (j *DbJob) Cancel(reason string) error {
	return j.updateTaskStatus(db.DbTaskStatusCancelled, reason)
}

func (j *DbJob) updateTaskStatus(status db.DbTaskStatus, reason string) error {
	conn, err := gd_.DbPool.Acquire(gd_.ConnCtx)
	if err != nil {
		j.Log().Error().Err(err).Str("NewStatus", string(status)).Msg("Failed to update task status")
		return err
	}
	defer conn.Release()

	q := db.New(conn)
	if err = q.SetDbTaskStatus(gd_.ConnCtx, db.SetDbTaskStatusParams{ID: j.ID(), Status: status, ErrReason: reason}); err != nil {
		j.Log().Error().Err(err).Str("NewStatus", string(status)).Msg("Failed to update task status")
		return err
	}

	return nil
}

func (j *DbJob) updateDbStatus(stage proto.DbStage, status proto.DbStatus) error {
	conn, err := gd_.DbPool.Acquire(gd_.ConnCtx)
	if err != nil {
		j.Log().Error().Err(err).Str("NewStatus", status.String()).Msg("Failed to update db status")
		return err
	}
	defer conn.Release()

	q := db.New(conn)
	if err = q.SetDbStatus(gd_.ConnCtx, db.SetDbStatusParams{ID: j.DbID, Status: status, Stage: stage}); err != nil {
		j.Log().Error().Err(err).Str("NewStatus", status.String()).Msg("Failed to update db status")
		return err
	}
	return nil
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
