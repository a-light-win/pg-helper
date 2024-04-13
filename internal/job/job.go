package job

import "github.com/google/uuid"

type Job interface {
	ID() uuid.UUID
	Name() string
	Requires() []uuid.UUID

	Run()
	Cancel(reason string) error

	IsPending() bool
	IsRunning() bool
	IsDone() bool
	IsFailed() bool
}
