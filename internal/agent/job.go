package agent

import (
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/a-light-win/pg-helper/internal/job/db_job"
	"github.com/rs/zerolog/log"
)

func (a *Agent) initJob() error {
	a.JobScheduler = job.NewJobScheduler(a.QuitCtx, 8)

	a.JobProducer = &job.JobProducer{
		AddJobs: a.JobScheduler.AddJobs,
	}

	return nil
}

func (a *Agent) runJob() {
	if jobs, err := db_job.RecoverJobs(); err != nil {
		log.Fatal().Err(err).Msg("Failed to recover jobs from database")
		return
	} else {
		a.JobScheduler.Init(jobs)
	}

	go a.JobScheduler.Run()
}
