package job

import "github.com/a-light-win/pg-helper/pkg/server"

type DoneJobHandler struct {
	PendingJobHandler *PendingJobHandler
}

func (h *DoneJobHandler) Handle(msg server.NamedElement) error {
	job := msg.(Job)
	h.PendingJobHandler.OnJobDone(job)
	return nil
}

func (h *DoneJobHandler) Init(setter server.GlobalSetter) error {
	return nil
}

func (h *DoneJobHandler) PostInit(getter server.GlobalGetter) error {
	return nil
}
