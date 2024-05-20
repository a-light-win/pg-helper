package job

import (
	"context"
	"fmt"
	"sync"

	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type PendingJob struct {
	Job        Job
	WaitingFor []uuid.UUID
}

type JobScheduler struct {
	// The key is the job that not done yet,
	// the value is the list of jobs that are waiting for the key job
	PendingJobs    map[uuid.UUID][]*PendingJob
	ReadytoRunJobs chan Job
	DoneJobs       chan Job
	AddJobs        chan Job
	QuitCtx        context.Context
	// Max number of concurrent jobs, the minimum value is 4
	MaxConcurrency int

	// For graceful shutdown
	wg sync.WaitGroup
	// Protects PendingJobs
	pendingLock sync.Mutex

	handler JobHandler
}

func NewJobScheduler(ctx context.Context, handler JobHandler, max int) *JobScheduler {
	scheduler := &JobScheduler{
		PendingJobs:    make(map[uuid.UUID][]*PendingJob),
		ReadytoRunJobs: make(chan Job),
		DoneJobs:       make(chan Job),
		AddJobs:        make(chan Job),
		QuitCtx:        ctx,
		handler:        handler,
		MaxConcurrency: max,
	}
	return scheduler
}

func (js *JobScheduler) InitJobs(jobs []Job) {
	js.pendingLock.Lock()
	defer js.pendingLock.Unlock()

	pendingJobs := []*PendingJob{}
	for _, job := range jobs {
		pendingJob := &PendingJob{Job: job}
		js.PendingJobs[job.ID()] = []*PendingJob{}

		if job.IsPending() {
			pendingJobs = append(pendingJobs, pendingJob)
		} else if job.IsRunning() {
			// We do not cancel the running jobs here,
			// It is the responsibility of the caller to handle the running jobs
			js.ReadytoRunJobs <- job
		}
	}

	for _, pendingJob := range pendingJobs {
		js.addJob(pendingJob)
	}
}

func (js *JobScheduler) AddJob(job Job) {
	if !job.IsPending() {
		log.Info().Msgf("Job %s is ignored because it is not in pending statues", job.Name())
		return
	}

	pendingJob := &PendingJob{
		Job:        job,
		WaitingFor: []uuid.UUID{},
	}

	js.pendingLock.Lock()
	defer js.pendingLock.Unlock()

	js.PendingJobs[job.ID()] = []*PendingJob{}

	js.addJob(pendingJob)
}

func (js *JobScheduler) addJob(pendingJob *PendingJob) {
	job := pendingJob.Job
	log.Debug().Msgf("Job %s is added to queue", job.Name())

	for _, requiredJobID := range job.Requires() {
		if pendingJobs, ok := js.PendingJobs[requiredJobID]; ok {
			pendingJob.WaitingFor = append(pendingJob.WaitingFor, requiredJobID)
			js.PendingJobs[requiredJobID] = append(pendingJobs, pendingJob)
		}
	}

	if pendingJob.WaitingFor == nil || len(pendingJob.WaitingFor) == 0 {
		js.ReadytoRunJobs <- job
	}
}

func (js *JobScheduler) OnJobDone(job Job) {
	if job.IsFailed() {
		log.Debug().Msgf("Job %s is failed", job.Name())
	} else {
		log.Debug().Msgf("Job %s is done", job.Name())
	}

	js.pendingLock.Lock()
	defer js.pendingLock.Unlock()

	if pendingJobs, ok := js.PendingJobs[job.ID()]; ok {
		for _, pendingJob := range pendingJobs {
			pendingJob.WaitingFor = removeUUID(pendingJob.WaitingFor, job.ID())
			if job.IsFailed() {
				reason := fmt.Sprintf("the dependency job %s is failed/canceled", job.Name())
				// TODO: How to recover if cancel failed?
				js.cancelJob(pendingJob.Job, reason)
			} else {
				if len(pendingJob.WaitingFor) == 0 {
					log.Debug().Msgf("Job %s is ready to run", pendingJob.Job.Name())
					js.ReadytoRunJobs <- pendingJob.Job
				}
			}
		}
		delete(js.PendingJobs, job.ID())
	}
}

func (js *JobScheduler) CancelJob(job Job, reason string) error {
	js.pendingLock.Lock()
	defer js.pendingLock.Unlock()

	return js.cancelJob(job, reason)
}

func (js *JobScheduler) cancelJob(job Job, reason string) error {
	log.Debug().Msgf("Job %s is canceld because %s", job.Name(), reason)
	if err := js.handler.Cancel(job, reason); err != nil {
		log.Error().Err(err).Str("Name", job.Name()).Msg("Cancel job failed")
		return err
	}
	js.DoneJobs <- job
	return nil
}

func removeUUID(uuids []uuid.UUID, target uuid.UUID) []uuid.UUID {
	for i, uuid := range uuids {
		if uuid == target {
			uuids = append(uuids[:i], uuids[i+1:]...)
			break
		}
	}
	return uuids
}

func (js *JobScheduler) Init(setter handler.GlobalSetter) error {
	if err := js.handler.Init(setter); err != nil {
		return err
	}

	setter.Set("job_producer", &JobProducer{AddJobs: js.AddJobs})

	return nil
}

func (js *JobScheduler) PostInit(getter handler.GlobalGetter) error {
	if err := js.handler.PostInit(getter); err != nil {
		return err
	}

	if jobs, err := js.handler.RecoverJobs(); err == nil {
		js.InitJobs(jobs)
	} else {
		log.Error().Err(err).Msg("Failed to recover jobs")
		return err
	}

	return nil
}

func (js *JobScheduler) Run() {
	log.Log().Msg("Job scheduler is running")

	sem := make(chan struct{}, max(js.MaxConcurrency, 4))
	defer close(sem)

	for {
		select {
		case <-js.QuitCtx.Done():
			return
		case job := <-js.AddJobs:
			js.AddJob(job)
		case job := <-js.ReadytoRunJobs:
			sem <- struct{}{}
			js.wg.Add(1)
			go func(job Job) {
				log.Debug().Msgf("Job %s is running", job.Name())

				if err := js.handler.Run(job); err != nil {
					log.Error().Err(err).Str("Name", job.Name()).Msg("Run job failed")
				}

				js.DoneJobs <- job
				<-sem
				js.wg.Done()
			}(job)
		case job := <-js.DoneJobs:
			js.OnJobDone(job)
		}
	}
}

func (js *JobScheduler) Shutdown(ctx context.Context) error {
	log.Log().Msg("Job scheduler is waiting for graceful shutdown")

	<-js.QuitCtx.Done()
	js.wg.Wait()

	log.Log().Msg("Job scheduler is shutdown gracefully")

	close(js.ReadytoRunJobs)
	close(js.DoneJobs)
	close(js.AddJobs)

	return nil
}
