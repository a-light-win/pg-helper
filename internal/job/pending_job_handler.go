package job

import (
	"fmt"
	"sync"

	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type PendingJob struct {
	Job
	WaitingFor []uuid.UUID
}

type PendingJobHandler struct {
	// THe key is the job that not done yet,
	// the value is the list of jobs that are waiting for the key job
	pendingJobs map[uuid.UUID][]*PendingJob
	pendingLock sync.Mutex

	readyToRunJobProducer handler.Producer
}

func (h *PendingJobHandler) Init(setter handler.GlobalSetter) error {
	h.pendingJobs = make(map[uuid.UUID][]*PendingJob)
	return nil
}

func (h *PendingJobHandler) PostInit(getter handler.GlobalGetter) error {
	h.readyToRunJobProducer = getter.Get(constants.AgentKeyReadyToRunJobProducer).(handler.Producer)

	// TODO: init jobs
	// dbApi := getter.Get(constants.AgentKeyDbApi).(proto.DbTaskSvcClient)

	return nil
}

func (h *PendingJobHandler) Handle(msg handler.NamedElement) error {
	job := msg.(Job)

	h.pendingLock.Lock()
	defer h.pendingLock.Unlock()

	h.addJob(&PendingJob{Job: job})
	return nil
}

func (h *PendingJobHandler) initJobs(jobs []Job) {
	h.pendingLock.Lock()
	defer h.pendingLock.Unlock()

	pendingJobs := []*PendingJob{}
	for _, job := range jobs {
		h.pendingJobs[job.UUID()] = []*PendingJob{}

		pendingJob := &PendingJob{Job: job}
		pendingJobs = append(pendingJobs, pendingJob)
	}

	log.Info().Int("Job Count", len(pendingJobs)).
		Msg("Load jobs from db")

	for _, pendingJob := range pendingJobs {
		h.addJob(pendingJob)
	}
}

func (h *PendingJobHandler) addJob(job *PendingJob) {
	if !job.IsPending() && !job.IsRunning() {
		log.Error().Str("JobName", job.GetName()).Msg("Job is ignored because it is not in pending statues")
		return
	}

	if _, ok := h.pendingJobs[job.UUID()]; !ok {
		h.pendingJobs[job.UUID()] = []*PendingJob{}
	}

	for _, requiredJobID := range job.Requires() {
		if pendingJobs, ok := h.pendingJobs[requiredJobID]; ok {
			h.pendingJobs[requiredJobID] = append(pendingJobs, job)
			// job is waiting for rqeuiredJobID
			job.WaitingFor = append(job.WaitingFor, requiredJobID)
		}
	}

	if job.WaitingFor == nil || len(job.WaitingFor) == 0 {
		log.Debug().Str("JobName", job.GetName()).Msg("Job is ready to run")
		h.readyToRunJobProducer.Send(job.Job)
	}
}

func (h *PendingJobHandler) OnJobDone(job Job) {
	log.Debug().Str("JobName", job.GetName()).Msgf("Job is done")

	h.pendingLock.Lock()
	defer h.pendingLock.Unlock()

	if pendingJobs, ok := h.pendingJobs[job.UUID()]; ok {
		for _, pendingJob := range pendingJobs {
			pendingJob.WaitingFor = removeUUID(pendingJob.WaitingFor, job.UUID())
			if job.IsFailed() {
				reason := fmt.Sprintf("the dependency job %s is failed/canceled", job.GetName())
				log.Debug().Msg(reason)
				pendingJob.Cancelling(reason)
				h.readyToRunJobProducer.Send(pendingJob.Job)
			} else if len(pendingJob.WaitingFor) == 0 {
				log.Debug().Msgf("Job %s is ready to run", pendingJob.GetName())
				h.readyToRunJobProducer.Send(pendingJob.Job)
			}
		}
		delete(h.pendingJobs, job.UUID())
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
