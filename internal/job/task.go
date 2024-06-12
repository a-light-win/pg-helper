package job

import (
	"sync"

	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/google/uuid"
)

type TaskStatus interface {
	IsPending() bool
	IsRunning() bool
	IsFailed() bool
	IsDone() bool
}

type SubTask interface {
	server.NamedElement

	TaskStatus

	UUID() uuid.UUID
	JobID() uuid.UUID

	Cancel(reason string)
}

type TaskDependency interface {
	IsReadyToRun() bool
	IsCancelling() bool
	CancelledBy() uuid.UUID

	// return true if the task is ready to run
	DependsDone(uuid.UUID) bool
	// return true if the task is ready to run
	DependsFailed(uuid.UUID) bool
	// Cancel the task if the depends failed
}

type Task interface {
	SubTask
	TaskDependency
}

type BaseTaskDependency struct {
	task TaskStatus

	liveDependsOn   []uuid.UUID
	liveDependsLock sync.Mutex

	cancelling bool
	cancelBy   uuid.UUID
}

func NewBaseTaskDependency(taskStatus TaskStatus, dependsOn []uuid.UUID) *BaseTaskDependency {
	return &BaseTaskDependency{task: taskStatus, liveDependsOn: dependsOn}
}

func (t *BaseTaskDependency) DependsDone(taskId uuid.UUID) bool {
	t.liveDependsLock.Lock()
	defer t.liveDependsLock.Unlock()

	if t.task.IsDone() {
		return false
	}

	if t.removeDependsOn(taskId) {
		return len(t.liveDependsOn) == 0
	}
	return false
}

func (t *BaseTaskDependency) DependsFailed(taskId uuid.UUID) bool {
	t.liveDependsLock.Lock()
	defer t.liveDependsLock.Unlock()

	if t.task.IsDone() {
		return false
	}
	if t.removeDependsOn(taskId) {
		t.cancelling = true
		t.cancelBy = taskId
		return true
	}
	return false
}

func (t *BaseTaskDependency) IsReadyToRun() bool {
	t.liveDependsLock.Lock()
	defer t.liveDependsLock.Unlock()

	if t.task.IsDone() {
		return false
	}
	return len(t.liveDependsOn) == 0 || t.cancelling
}

func (t *BaseTaskDependency) removeDependsOn(taskId uuid.UUID) bool {
	for i, id := range t.liveDependsOn {
		if id == taskId {
			t.liveDependsOn = append(t.liveDependsOn[:i], t.liveDependsOn[i+1:]...)
			return true
		}
	}
	return false
}

func (t *BaseTaskDependency) IsCancelling() bool {
	return !t.task.IsDone() && t.cancelling
}

func (t *BaseTaskDependency) CancelledBy() uuid.UUID {
	return t.cancelBy
}
