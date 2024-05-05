package grpc_server

import (
	"errors"
	"sort"
	"sync"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/rs/zerolog"
)

type DbInstanceManager struct {
	Instances map[string]*DbInstance

	instLock sync.Mutex
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

	inst := NewDbInstance(instName, pgVersion, logger)
	m.addInstance(inst)
	return inst, nil
}

func (m *DbInstanceManager) addInstance(inst *DbInstance) {
	m.Instances[inst.Name] = inst
}

type InstanceFilter struct {
	// The Instance Name
	Name string `validate:"ommitempty,max=63,id"`
	// The Postgres major version
	Version int32 `validate:"omitempty,pg_ver"`
	// The database name
	DbName string `validate:"max=63,id"`
	// Database must be in the instance
	MustExist bool `validate:"omitempty"`
}

func (m *DbInstanceManager) FilterInstances(filter *InstanceFilter) []*DbInstance {
	m.instLock.Lock()
	defer m.instLock.Unlock()

	return m.filterInstances(filter)
}

func (m *DbInstanceManager) FirstMatchedInstance(filter *InstanceFilter) *DbInstance {
	m.instLock.Lock()
	defer m.instLock.Unlock()

	instances := m.filterInstances(filter)
	if len(instances) > 0 {
		return instances[0]
	}
	return nil
}

func (m *DbInstanceManager) filterInstances(filter *InstanceFilter) []*DbInstance {
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
			if db != nil && db.Stage != proto.DbStage_MigrateOut && db.Stage != proto.DbStage_Dropping && db.Stage != proto.DbStage_None {
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
