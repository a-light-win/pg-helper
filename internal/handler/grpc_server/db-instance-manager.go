package grpc_server

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	handler "github.com/a-light-win/pg-helper/internal/interface/grpc_server"
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

func (m *DbInstanceManager) FilterInstances(filter *handler.InstanceFilter) []*DbInstance {
	m.instLock.Lock()
	defer m.instLock.Unlock()

	return m.filterInstances(filter)
}

func (m *DbInstanceManager) FirstMatchedInstance(filter *handler.InstanceFilter) *DbInstance {
	instances := m.FilterInstances(filter)
	if len(instances) > 0 {
		return instances[0]
	}
	return nil
}

func (m *DbInstanceManager) filterInstances(filter *handler.InstanceFilter) []*DbInstance {
	var result []*DbInstance
	matched := false

	for _, inst := range m.Instances {
		if filter.Name != "" {
			if inst.Name == filter.Name {
				matched = true
			} else {
				continue
			}
		}

		if filter.Version != 0 && inst.PgVersion != filter.Version {
			continue
		}

		if filter.DbName != "" {
			db := inst.GetDb(filter.DbName)
			if db == nil && filter.MustExist {
				continue
			}
			if db != nil &&
				db.Stage != proto.DbStage_Idle &&
				db.Stage != proto.DbStage_Dropping &&
				db.Stage != proto.DbStage_DropCompleted &&
				db.Stage != proto.DbStage_None {
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

func (m *DbInstanceManager) IsDbReady(request *handler.DbRequest) bool {
	inst := m.FirstMatchedInstance(&request.InstanceFilter)
	if inst == nil {
		return false
	}
	return inst.IsDbReady(request.DbName)
}

func (m *DbInstanceManager) GetDbStatus(request *handler.DbRequest) (*handler.DbStatusResponse, error) {
	inst := m.FirstMatchedInstance(&request.InstanceFilter)
	if inst == nil {
		return nil, errors.New("instance not found")
	}
	db := inst.GetDb(request.DbName)
	if db == nil {
		return nil, errors.New("database not found")
	}
	return &handler.DbStatusResponse{
		InstanceName: inst.Name,
		Version:      inst.PgVersion,
		Name:         db.Name,
		Stage:        db.Stage.String(),
		Status:       db.Status.String(),
		UpdatedAt:    db.UpdatedAt.AsTime(),
	}, nil
}

func (m *DbInstanceManager) GetDb(vo *handler.DbRequest) (*proto.Database, error) {
	inst := m.FirstMatchedInstance(&vo.InstanceFilter)
	if inst == nil {
		return nil, errors.New("instance not found")
	}
	db := inst.GetDb(vo.DbName)
	if db == nil {
		return nil, errors.New("database not found")
	}
	return db.Database, nil
}

func (m *DbInstanceManager) CreateDb(request *handler.CreateDbRequest, waitReady bool) (*handler.DbStatusResponse, error) {
	inst := m.FirstMatchedInstance(&request.InstanceFilter)
	if inst == nil {
		return nil, errors.New("instance not found")
	}
	if request.MigrateFrom != "" {
		if m.GetInstance(request.MigrateFrom) == nil {
			return nil, errors.New("instance migrate from not found")
		}
	}

	db, err := inst.CreateDb(request)
	if err != nil {
		return nil, err
	}

	if !waitReady {
		return nil, nil
	}

	if db.Stage != proto.DbStage_Ready {
		m.waitReady(inst.Name, request.DbName)
	}

	response := &handler.DbStatusResponse{
		InstanceName: inst.Name,
		Version:      inst.PgVersion,

		Name:   db.Name,
		Stage:  db.Stage.String(),
		Status: db.Status.String(),

		UpdatedAt: db.UpdatedAt.AsTime(),
	}
	return response, nil
}

func (m *DbInstanceManager) SubscribeDbStatus(callback handler.SubscribeDbStatusFunc) {
	m.dbSubscriber.Subscribe(callback)
}

func (m *DbInstanceManager) SubscribeInstanceStatus(callback handler.SubscribeInstanceStatusFunc) {
	m.InstSubscriber.Subscribe(callback)
}

func (m *DbInstanceManager) waitReady(instName, dbName string) {
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	m.dbSubscriber.Subscribe(func(dbStatus *handler.DbStatusResponse) bool {
		if timeoutCtx.Err() != nil {
			return false
		}
		if dbStatus.Stage == "Ready" && dbStatus.Name == dbName && dbStatus.InstanceName == instName {
			cancel()
			return true
		}
		return false
	})
	<-timeoutCtx.Done()
}
