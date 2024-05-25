package grpc_server

import (
	"sync"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/pkg/handler"
)

type DbStatusSubscriber struct {
	subscribers      []handler.SubscribeDbStatusFunc
	subscribersMutex sync.Mutex
}

func (s *DbStatusSubscriber) Subscribe(subscriber handler.SubscribeDbStatusFunc) {
	s.subscribersMutex.Lock()
	defer s.subscribersMutex.Unlock()
	s.subscribers = append(s.subscribers, subscriber)
}

func (s *DbStatusSubscriber) OnDbStatusChanged(instance *DbInstance, db *Database) {
	if db.Stage != proto.DbStage_Ready &&
		db.Stage != proto.DbStage_Idle &&
		db.Stage != proto.DbStage_DropCompleted &&
		db.Status != proto.DbStatus_Failed &&
		db.Status != proto.DbStatus_Expired &&
		db.Status != proto.DbStatus_Cancelled {
		return
	}

	s.subscribersMutex.Lock()
	defer s.subscribersMutex.Unlock()

	if len(s.subscribers) == 0 {
		return
	}

	dbStatus := &handler.DbStatusResponse{
		Name:         db.Name,
		Stage:        db.Stage.String(),
		Status:       db.Status.String(),
		UpdatedAt:    db.UpdatedAt.AsTime(),
		InstanceName: instance.Name,
		Version:      instance.PgVersion,
	}

	s.notifyDbStatusChanged(dbStatus)
}

func (s *DbStatusSubscriber) notifyDbStatusChanged(dbStatus *handler.DbStatusResponse) {
	for i := 0; i < len(s.subscribers); i++ {
		if !s.subscribers[i](dbStatus) {
			s.subscribers = append(s.subscribers[:i], s.subscribers[i+1:]...)
			i--
		}
	}
}
