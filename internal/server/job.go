package server

import (
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/rs/zerolog/log"
)

func (s *Server) initJob() error {
	s.JobScheduler = job.NewJobScheduler(s.QuitCtx, 8)
	// Initialize the job producer.
	// s.JobProducer = job.NewJobProducer(s.DbPool, &s.Config.Db, s.JobScheduler.AddJobs)

	if jobs, err := s.JobProducer.RecoverJobs(); err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize the job producer")
		return err
	} else {
		s.JobScheduler.Init(jobs)
	}

	return nil
}

func (s *Server) runJobScheduler() {
	go s.JobScheduler.Run()
}

func (s *Server) WaitJobSchedulerExit() {
	s.JobScheduler.WaitGracefulShutdown()
}
