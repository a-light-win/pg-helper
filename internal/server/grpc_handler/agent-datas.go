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

func (a *AgentDatas) AddAgent(agentId string, pgVersion int32, s proto.DbTaskSvc_RegisterServer) {
	agent := &AgentData{
		ID:         agentId,
		PgVersion:  pgVersion,
		Databases:  make(map[string]*Database),
		TaskSender: s,
	}
	a.Agents = append(a.Agents, agent)
}
