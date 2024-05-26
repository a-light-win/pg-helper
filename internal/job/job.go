package job

import (
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/google/uuid"
)

type Job interface {
	server.NamedElement

	UUID() uuid.UUID
	IsPending() bool
	IsRunning() bool
	IsCancelling() bool
	IsFailed() bool

	Requires() []uuid.UUID
	Cancelling(reason string)
}

type PendingJob struct {
	Job
	LiveDependsOn []uuid.UUID
	RequiredBy    []uuid.UUID
}

type InitJobProvider interface {
	RecoverJobs() (jobs []Job, err error)
}
