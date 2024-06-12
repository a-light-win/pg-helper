package job

import (
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/google/uuid"
)

type Job interface {
	server.NamedElement

	UUID() uuid.UUID

	Init()
	// return ready to run tasks
	OnTaskDone(task Task) []Task
	ReadyToRun() []Task

	IsFailed() bool
	IsDone() bool
}

type BaseJob struct {
	Name string
	ID   uuid.UUID

	Tasks []Task
}

func (j *BaseJob) GetName() string {
	if j.Name != "" {
		return j.Name
	}
	return j.ID.String()
}

func (j *BaseJob) UUID() uuid.UUID {
	return j.ID
}

func (j *BaseJob) IsFailed() bool {
	for _, task := range j.Tasks {
		if task.IsFailed() {
			return true
		}
	}
	return false
}

func (j *BaseJob) IsDone() bool {
	for _, task := range j.Tasks {
		if !task.IsDone() {
			return false
		}
	}
	return true
}

func (j *BaseJob) Init() {
	for _, task := range j.Tasks {
		if task.IsDone() {
			j.OnTaskDone(task)
		}
	}
}

func (j *BaseJob) OnTaskDone(doneTask Task) []Task {
	var tasks []Task
	if doneTask.IsFailed() {
		for _, task := range j.Tasks {
			if task.DependsFailed(doneTask.UUID()) {
				tasks = append(tasks, task)
			}
		}
	} else {
		for _, task := range j.Tasks {
			if task.DependsDone(doneTask.UUID()) {
				tasks = append(tasks, task)
			}
		}
	}
	return tasks
}

func (j *BaseJob) ReadyToRun() []Task {
	var tasks []Task
	for _, task := range j.Tasks {
		if task.IsReadyToRun() {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

type InitJobProvider interface {
	RecoverJobs() (jobs []Job, err error)
}
