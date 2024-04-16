package grpc_handler

import (
	"github.com/a-light-win/pg-helper/api/proto"
)

type AgentDatas struct {
	Agents []*AgentData
}

func (a *AgentDatas) GetAgent(id string) *AgentData {
	for _, agent := range a.Agents {
		if agent.ID == id {
			return agent
		}
	}
	return nil
}

func (a *AgentDatas) AddAgent(m *proto.RegisterAgent, s proto.DbTaskSvc_RegisterServer) {
	agent := &AgentData{
		ID:         m.AgentId,
		PgVersion:  m.PgVersion,
		Databases:  make(map[string]*Database),
		TaskSender: s,
	}
	a.Agents = append(a.Agents, agent)
}
