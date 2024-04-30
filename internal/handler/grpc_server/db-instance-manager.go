package grpc_server

import (
	"sync"

	"github.com/rs/zerolog/log"
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

func (m *DbInstanceManager) NewInstance(instName string, pgVersion int32) *DbInstance {
	m.instLock.Lock()
	defer m.instLock.Unlock()

	if inst := m.instance(instName); inst != nil {
		if inst.PgVersion != pgVersion {

			log.Log().
				Str("Agent", instName).
				Int32("OldPgVersion", inst.PgVersion).
				Int32("NewPgVersion", pgVersion).
				Msg("Agent updates pg version")

			inst.PgVersion = pgVersion
		}
		return inst
	}

	inst := NewDbInstance(instName, pgVersion)
	m.addInstance(inst)
	return inst
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
