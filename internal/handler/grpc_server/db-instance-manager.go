package grpc_server

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	api "github.com/a-light-win/pg-helper/internal/interface/grpcServerApi"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/rs/zerolog"
)

type DbInstanceManager struct {
	Instances map[string]*DbInstance
	instLock  sync.Mutex

	dbSubscriber   *DbStatusSubscriber
	InstSubscriber *InstanceStatusSubscriber
}

func NewDbInstanceManager() *DbInstanceManager {
	return &DbInstanceManager{
		Instances:      make(map[string]*DbInstance),
		dbSubscriber:   &DbStatusSubscriber{},
		InstSubscriber: &InstanceStatusSubscriber{},
	}
}

func (m *DbInstanceManager) GetInstance(instName string) *DbInstance {
	m.instLock.Lock()
	defer m.instLock.Unlock()

	return m.instance(instName)
}

func (m *DbInstanceManager) instance(instName string) *DbInstance {
	if inst, ok := m.Instances[instName]; ok {
		return inst
	}
	return nil
}

func (m *DbInstanceManager) NewInstance(instName string, pgVersion int32, logger *zerolog.Logger) (*DbInstance, error) {
	m.instLock.Lock()
	defer m.instLock.Unlock()

	if inst := m.instance(instName); inst != nil {
		if inst.PgVersion != pgVersion {
			err := errors.New("change pg instance version is not supported")
			logger.Error().Int32("OldPgVersion", inst.PgVersion).
				Err(err).Msg("New pg instance failed")
			return nil, err
		}
		inst.logger = logger
		return inst, nil
	}

	inst := NewDbInstance(instName, pgVersion, logger, m.dbSubscriber)
	m.addInstance(inst)
	return inst, nil
}

func (m *DbInstanceManager) addInstance(inst *DbInstance) {
	m.Instances[inst.Name] = inst
}

func (m *DbInstanceManager) FilterInstances(filter *api.InstanceFilter) []*DbInstance {
	m.instLock.Lock()
	defer m.instLock.Unlock()

	return m.filterInstances(filter)
}

func (m *DbInstanceManager) FirstMatchedInstance(filter *api.InstanceFilter) *DbInstance {
	instances := m.FilterInstances(filter)
	if len(instances) > 0 {
		return instances[0]
	}
	return nil
}

func (m *DbInstanceManager) filterInstances(filter *api.InstanceFilter) []*DbInstance {
	var result []*DbInstance
	matched := false

	for _, inst := range m.Instances {
		if filter.InstanceName != "" {
			if inst.Name == filter.InstanceName {
				matched = true
			} else {
				continue
			}
		}

		if filter.Version != 0 && inst.PgVersion != filter.Version {
			continue
		}

		if filter.Name != "" {
			db := inst.GetDb(filter.Name)
			if db == nil && filter.MustExist {
				continue
			}
			if db != nil &&
				!db.IsAlreadyIdle() &&
				!db.IsNotExist() {
				matched = true
			}
		}

		if matched {
			result = append([]*DbInstance{inst}, result...)
			break
		}
		result = append(result, inst)
	}

	if !matched && filter.Version == 0 {
		// Sort the result by instance's version desc
		sort.Slice(result, func(i, j int) bool {
			return result[i].PgVersion > result[j].PgVersion
		})
	}

	return result
}

func (m *DbInstanceManager) GetDbStatus(request *api.DbRequest) (*api.DbStatusResponse, error) {
	inst := m.FirstMatchedInstance(&request.InstanceFilter)
	if inst == nil {
		return nil, errors.New("instance not found")
	}
	db := inst.GetDb(request.Name)
	if db == nil {
		return nil, errors.New("database not found")
	}
	return &api.DbStatusResponse{
		InstanceName: inst.Name,
		Version:      inst.PgVersion,
		Name:         db.Name,
		Stage:        db.Stage.String(),
		Status:       db.Status.String(),
		UpdatedAt:    db.UpdatedAt.AsTime(),
	}, nil
}

func (m *DbInstanceManager) GetDb(vo *api.DbRequest) (*proto.Database, error) {
	inst := m.FirstMatchedInstance(&vo.InstanceFilter)
	if inst == nil {
		return nil, errors.New("instance not found")
	}
	db := inst.GetDb(vo.Name)
	if db == nil {
		return nil, errors.New("database not found")
	}
	return db.Database, nil
}

func (m *DbInstanceManager) CreateDb(request *api.CreateDbRequest) error {
	inst := m.FirstMatchedInstance(&request.InstanceFilter)
	if inst == nil || !inst.Online {
		return api.ErrInstanceOffline
	}

	if request.MigrateFrom != "" {
		oldInst := m.GetInstance(request.MigrateFrom)
		if oldInst == nil || !oldInst.Online {
			return api.ErrInstanceOffline
		}

		migrateOutRequest := &api.MigrateOutDbRequest{
			Name:         request.Name,
			InstanceName: request.MigrateFrom,
			Reason:       request.Reason,
			MigrateTo:    inst.Name,
		}
		return oldInst.MigrateOut(migrateOutRequest,
			func() error { return inst.CreateDb(request) })
	}

	return inst.CreateDb(request)
}

func (m *DbInstanceManager) SubscribeDbStatus(callback api.SubscribeDbStatusFunc) {
	m.dbSubscriber.Subscribe(callback)
}

func (m *DbInstanceManager) SubscribeInstanceStatus(callback api.SubscribeInstanceStatusFunc) {
	m.InstSubscriber.Subscribe(callback)
}

func (m *DbInstanceManager) WaitReady(instName, dbName string, timeout time.Duration) bool {
	if m.isReady(instName, dbName) {
		return true
	}

	timeoutCtx, cancel := context.WithTimeout(context.Background(), timeout)
	ready := false
	m.dbSubscriber.Subscribe(func(dbStatus *api.DbStatusResponse) bool {
		if timeoutCtx.Err() != nil {
			return api.StopSubscribe
		}
		if dbStatus.IsReady(dbName, instName) {
			ready = true
			cancel()
			return api.StopSubscribe
		}
		return api.ContinueSubscribe
	})
	<-timeoutCtx.Done()
	return ready
}

func (m *DbInstanceManager) isReady(instName, dbName string) bool {
	inst := m.GetInstance(instName)
	if inst == nil {
		return false
	}
	db := inst.GetDb(dbName)
	if db == nil {
		return false
	}
	return db.IsReadyToUse()
}
