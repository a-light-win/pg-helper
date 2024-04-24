package grpc_server

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type AgentDatas struct {
	Agents    []*AgentData
	agentLock sync.Mutex
}

func (a *AgentDatas) GetAgent(id string) *AgentData {
	a.agentLock.Lock()
	defer a.agentLock.Unlock()

	return a.getAgent(id)
}

func (a *AgentDatas) getAgent(id string) *AgentData {
	for _, agent := range a.Agents {
		if agent.ID == id {
			return agent
		}
	}
	return nil
}

func (a *AgentDatas) NewAgent(agentId string, pgVersion int32) *AgentData {
	a.agentLock.Lock()
	defer a.agentLock.Unlock()

	if agent := a.getAgent(agentId); agent != nil {
		if agent.PgVersion != pgVersion {

			log.Log().
				Str("Agent", agentId).
				Int32("OldPgVersion", agent.PgVersion).
				Int32("NewPgVersion", pgVersion).
				Msg("Agent updates pg version")

			agent.PgVersion = pgVersion
		}
		return agent
	}

	agent := NewAgentData(agentId, pgVersion)
	a.Agents = append(a.Agents, agent)
	return agent
}
