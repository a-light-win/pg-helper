package job

import (
	"fmt"
	"sync"

	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type PendingJobHandler struct {
	// THe key is the job that not done yet,
	// the value is the list of jobs that are waiting for the key job
	pendingJobs map[uuid.UUID]*PendingJob
	pendingLock sync.Mutex

	readyToRunJobProducer server.Producer
	doneJobProducer       server.Producer
	InitJobProvider       InitJobProvider
}

func (h *PendingJobHandler) Init(setter server.GlobalSetter) error {
	h.pendingJobs = make(map[uuid.UUID]*PendingJob)
	return nil
}

func (h *PendingJobHandler) PostInit(getter server.GlobalGetter) error {
	h.readyToRunJobProducer = getter.Get(constants.AgentKeyReadyToRunJobProducer).(server.Producer)
	h.doneJobProducer = getter.Get(constants.AgentKeyDoneJobProducer).(server.Producer)

	if jobs, err := h.InitJobProvider.RecoverJobs(); err != nil {
		log.Error().Err(err).Msg("Failed to recover jobs")
		return err
	} else {
		h.InitJobs(jobs)
	}

	return nil
}

func (h *PendingJobHandler) Handle(msg server.NamedElement) error {
	job := msg.(Job)

	h.pendingLock.Lock()
	defer h.pendingLock.Unlock()

	h.addJob(&PendingJob{Job: job})
	return nil
}

func (h *PendingJobHandler) InitJobs(jobs []Job) {
	h.pendingLock.Lock()
	defer h.pendingLock.Unlock()

	for _, job := range jobs {
		h.pendingJobs[job.UUID()] = &PendingJob{Job: job}
	}

	log.Info().Int("Job Count", len(h.pendingJobs)).
		Msg("Load jobs from db")

	for _, job := range h.pendingJobs {
		h.initDepends(job)
		h.checkReadyToRun(job)
	}
}

func (h *PendingJobHandler) initDepends(job *PendingJob) {
	for _, requiredJobID := range job.Requires() {
		if pendingJob, ok := h.pendingJobs[requiredJobID]; ok {
			job.LiveDependsOn = append(job.LiveDependsOn, pendingJob.UUID())
			pendingJob.RequiredBy = append(pendingJob.RequiredBy, job.UUID())
		}
	}
}

func (h *PendingJobHandler) checkReadyToRun(job *PendingJob) {
	if job.LiveDependsOn == nil || len(job.LiveDependsOn) == 0 || job.IsFailed() {
		log.Debug().Str("JobName", job.GetName()).Msg("Job is ready to run")
		h.readyToRunJobProducer.Send(job.Job)
	}
}

func (h *PendingJobHandler) addJob(job *PendingJob) {
	if !job.IsPending() && !job.IsRunning() {
		log.Error().Str("JobName", job.GetName()).Msg("Job is ignored because it is not in pending statues")
		if pendingJob, ok := h.pendingJobs[job.UUID()]; ok {
			h.removeJob(pendingJob.UUID())
		}
		return
	}

	if _, ok := h.pendingJobs[job.UUID()]; !ok {
		log.Debug().Str("JobName", job.GetName()).Msg("Job is added")
		h.pendingJobs[job.UUID()] = job
		h.initDepends(job)
		h.checkReadyToRun(job)
	} else {
		log.Warn().Str("JobName", job.GetName()).Msg("Job already exists")
	}
}

func (h *PendingJobHandler) OnJobDone(job Job) {
	log.Debug().Str("JobName", job.GetName()).Msgf("Job is done")

	h.pendingLock.Lock()
	defer h.pendingLock.Unlock()

	h.removeJob(job.UUID())
}

func (h *PendingJobHandler) removeJob(jobId uuid.UUID) {
	if job, ok := h.pendingJobs[jobId]; ok {
		for _, requiredByID := range job.RequiredBy {
			if pendingJob, ok := h.pendingJobs[requiredByID]; ok {
				pendingJob.LiveDependsOn = removeUUID(pendingJob.LiveDependsOn, jobId)
				if job.IsFailed() {
					reason := fmt.Sprintf("the dependency job %s is failed/canceled", job.GetName())

					log.Debug().Str("JobName", pendingJob.GetName()).
						Str("Reason", reason).
						Msg("Job is canceled")

					pendingJob.Cancelling(reason)
				}
				h.checkReadyToRun(pendingJob)
			}
		}
		log.Debug().Str("JobName", job.GetName()).Msg("Job is removed")
		delete(h.pendingJobs, jobId)
	} else {
		log.Debug().Str("JobName", job.GetName()).Msg("Job is not found")
	}
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
