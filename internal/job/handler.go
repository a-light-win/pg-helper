package job

import "github.com/a-light-win/pg-helper/pkg/handler"

type JobHandler interface {
	handler.Initialization

	Run(job Job) error
	Cancel(job Job, reason string) error

	RecoverJobs() (jobs []Job, err error)
}
