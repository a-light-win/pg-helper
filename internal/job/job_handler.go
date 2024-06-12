package job

import (
	"errors"
	"fmt"
	"sync"

	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/a-light-win/pg-helper/pkg/utils/logger"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type JobHandler struct {
	jobs     map[uuid.UUID]Job
	jobsLock sync.Mutex

	readyToRunJobProducer server.Producer
}

var (
	ErrUnknownMessageType     = errors.New("unknown message type")
	ErrJobAlreadyExist        = errors.New("job is already exist")
	ErrJobNotFound            = errors.New("job is not found")
	ErrNotSupportedTaskStatus = errors.New("not supported task status")
)

func (h *JobHandler) Init(setter server.GlobalSetter) error {
	h.jobs = make(map[uuid.UUID]Job)
	return nil
}

func (h *JobHandler) PostInit(getter server.GlobalGetter) error {
	h.readyToRunJobProducer = getter.Get(constants.AgentKeyReadyToRunJobProducer).(server.Producer)

	return nil
}

func (h *JobHandler) Handle(msg server.NamedElement) error {
	switch msg := msg.(type) {
	case Job:
		return h.addJob(msg)
	case Task:
		return h.processTask(msg)
	default:
		err := ErrUnknownMessageType
		log.Warn().Err(err).
			Interface("msg", msg).
			Msg("Unknown message type")
		return logger.NewAlreadyLoggedError(err, zerolog.WarnLevel)

	}
}

func (h *JobHandler) checkReadyToRun(job Job) {
	readyToRunTasks := job.ReadyToRun()
	h.readyToRun(job.GetName(), readyToRunTasks)
}

func (h *JobHandler) readyToRun(jobName string, readyToRunTasks []Task) {
	if len(readyToRunTasks) == 0 {
		return
	}
	for _, task := range readyToRunTasks {
		log.Debug().
			Str("JobName", jobName).
			Str("JobID", task.JobID().String()).
			Str("TaskName", task.GetName()).
			Msg("Task is ready to run")
		h.readyToRunJobProducer.Send(task)
	}
}

func (h *JobHandler) addJob(job Job) error {
	h.jobsLock.Lock()
	defer h.jobsLock.Unlock()

	if _, ok := h.jobs[job.UUID()]; ok {
		err := ErrJobAlreadyExist
		log.Warn().Err(err).
			Str("JobName", job.GetName()).
			Msg("Nothing to do")
		return logger.NewAlreadyLoggedError(err, zerolog.WarnLevel)
	}

	h.jobs[job.UUID()] = job

	h.checkReadyToRun(job)
	return nil
}

func (h *JobHandler) processTask(task Task) error {
	var err error
	if task.IsCancelling() {
		err = h.processCancellingTask(task)
		if err != nil {
			return err
		}
	}

	if task.IsDone() {
		err = h.processDoneTask(task)
	} else {
		err = ErrNotSupportedTaskStatus
	}

	log.Warn().Err(err).
		Str("JobID", task.JobID().String()).
		Str("TaskName", task.GetName()).
		Msg("Nothing to do")
	return logger.NewAlreadyLoggedError(err, zerolog.WarnLevel)
}

func (h *JobHandler) processDoneTask(task Task) error {
	h.jobsLock.Lock()
	defer h.jobsLock.Unlock()

	jobId := task.JobID()
	if job, ok := h.jobs[jobId]; ok {
		log.Debug().
			Str("JobName", job.GetName()).
			Str("JobID", task.JobID().String()).
			Str("TaskName", task.GetName()).
			Msg("Task is done")

		readyToRunTasks := job.OnTaskDone(task)
		h.readyToRun(job.GetName(), readyToRunTasks)
		if job.IsDone() {
			h.onJobDone(job)
		}
		return nil
	}
	return ErrJobNotFound
}

func (h *JobHandler) processCancellingTask(task Task) error {
	task.Cancel(fmt.Sprintf("Task is cancelled by %s", task.CancelledBy().String()))
	return nil
}

func (h *JobHandler) onJobDone(job Job) {
	log.Debug().
		Str("JobName", job.GetName()).
		Msgf("Job is done")

	delete(h.jobs, job.UUID())
}
