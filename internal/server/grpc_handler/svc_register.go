package grpc_handler

import (
	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/rs/zerolog/log"
)

func (h *DbTaskSvcHandler) Register(m *proto.RegisterAgent, s proto.DbTaskSvc_RegisterServer) error {
	authInfo, err := authInfoWithType(s.Context(), AgentClient)
	if err != nil {
		return err
	}

	agent := gd_.GetAgent(authInfo.ClientId)
	if agent == nil {

		log.Log().
			Str("AgentId", authInfo.ClientId).
			Int32("PgVersion", m.PgVersion).
			Msg("Agent register first time.")

		gd_.AddAgent(m, s)
	} else {
		log.Debug().
			Str("AgentId", authInfo.ClientId).
			Int32("PgVersion", m.PgVersion).
			Msg("Agent register again")
	}

	for _, db := range m.Databases {
		agent.UpdateDatabase(db)
	}
	return nil
}
