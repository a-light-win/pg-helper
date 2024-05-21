package job

import (
	"github.com/a-light-win/pg-helper/pkg/handler"
)

type DoneJobHandler struct {
	PendingJobHandler *PendingJobHandler
}

func (h *DoneJobHandler) Handle(msg handler.NamedElement) error {
	job := msg.(Job)
	h.PendingJobHandler.OnJobDone(job)
	return nil
}

func (h *DoneJobHandler) Init(setter handler.GlobalSetter) error {
	return nil
}

func (h *DoneJobHandler) PostInit(getter handler.GlobalGetter) error {
	return nil
}
