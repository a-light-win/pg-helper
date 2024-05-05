package grpc_server

import (
	"errors"
	"sync"

	"github.com/rs/zerolog"
)

type DbInstanceManager struct {
	Instances map[string]*DbInstance

	sortedInstances []*DbInstance
	instLock        sync.Mutex
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

func (m *DbInstanceManager) instancesByVersion(pgVersion int32) []*DbInstance {
	var result []*DbInstance
	for _, inst := range m.Instances {
		if inst.PgVersion == pgVersion {
			result = append(result, inst)
		}
	}
	return result
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

	// Sort instances by pg version,
	// the larger the version, the earlier in the slice
	for i, sortedInst := range m.sortedInstances {
		if sortedInst.PgVersion < inst.PgVersion {
			m.sortedInstances = append(m.sortedInstances[:i], append([]*DbInstance{inst}, m.sortedInstances[i:]...)...)
			return
		}
	}
	m.sortedInstances = append(m.sortedInstances, inst)
}
