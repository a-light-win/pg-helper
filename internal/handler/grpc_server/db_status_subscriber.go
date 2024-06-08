package grpc_server

import (
	"sync"

	api "github.com/a-light-win/pg-helper/internal/interface/grpcServerApi"
)

type DbStatusSubscriber struct {
	subscribers []api.SubscribeDbStatusFunc
	mutex       sync.Mutex
}

func (s *DbStatusSubscriber) Subscribe(subscriber api.SubscribeDbStatusFunc) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.subscribers = append(s.subscribers, subscriber)
}

func (s *DbStatusSubscriber) OnStatusChanged(instance *DbInstance, db *Database) {
	if !db.IsFailed() && !db.IsSynced() {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.subscribers) == 0 {
		return
	}

	dbStatus := db.StatusResponse()
	dbStatus.InstanceName = instance.Name
	dbStatus.Version = instance.PgVersion

	s.notifyStatusChanged(dbStatus)
}

func (s *DbStatusSubscriber) notifyStatusChanged(dbStatus *api.DbStatusResponse) {
	for i := 0; i < len(s.subscribers); i++ {
		if !s.subscribers[i](dbStatus) {
			s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
			i--
		}
	}
}
