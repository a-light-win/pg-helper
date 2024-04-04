package job

import (
	"context"

	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobProducer struct {
	DbPool *pgxpool.Pool
	Config *config.DbConfig
	Ctx    context.Context

	AddJobs chan Job
}

func NewJobProducer(dbPool *pgxpool.Pool, config *config.DbConfig, addJobs chan Job) *JobProducer {
	return &JobProducer{
		DbPool:  dbPool,
		Config:  config,
		Ctx:     context.Background(),
		AddJobs: addJobs,
	}
}

func (j *JobProducer) RecoverJobs() ([]Job, error) {
	conn, err := j.DbPool.Acquire(j.Ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	q := db.New(conn)
	activeTasks, err := q.ListActiveDbTasks(j.Ctx)
	if err != nil {
		if err != pgx.ErrNoRows {
			return nil, err
		}
		return nil, nil
	} else {
		jobs := make([]Job, 0, len(activeTasks))
		for _, task := range activeTasks {
			jobs = append(jobs, NewDbJob(j.Ctx, &task, j.Config, j.DbPool))
		}
		return jobs, nil
	}
}

func (j *JobProducer) Produce(task *db.DbTask) (readyChan chan db.DbTaskStatus) {
	job := NewDbJob(j.Ctx, task, j.Config, j.DbPool)
	readyChan = job.ReadyChan
	j.AddJobs <- job
	return
}
