package grpc_server

import (
	"sync"

	api "github.com/a-light-win/pg-helper/internal/interface/grpcServerApi"
)

type InstanceStatusSubscriber struct {
	subscribers []api.SubscribeInstanceStatusFunc
	mutex       sync.Mutex
}

func (s *InstanceStatusSubscriber) Subscribe(subscriber api.SubscribeInstanceStatusFunc) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.subscribers = append(s.subscribers, subscriber)
}

func (s *InstanceStatusSubscriber) OnStatusChanged(instance *DbInstance) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.subscribers) == 0 {
		return
	}

	instanceStatus := instance.StatusResponse()
	s.notifyStatusChanged(instanceStatus)
}

func (s *InstanceStatusSubscriber) notifyStatusChanged(instanceStatus *api.InstanceStatusResponse) {
	for i := 0; i < len(s.subscribers); i++ {
		if !s.subscribers[i](instanceStatus) {
			s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
			i--
		}
	}
}
