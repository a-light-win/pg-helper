package job

import (
	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/google/uuid"
)

type Job interface {
	handler.NamedElement

	UUID() uuid.UUID
	IsPending() bool
	IsRunning() bool
	IsFailed() bool

	Requires() []uuid.UUID
	Cancelling(reason string)
}

type PendingJob struct {
	Job
	WaitingFor []uuid.UUID
}

type InitJobProvider interface {
	RecoverJobs() (jobs []Job, err error)
}
