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

	agent := gd_.NewAgent(authInfo.Subject, m.PgVersion)
	log.Log().
		Str("AgentId", authInfo.Subject).
		Int32("PgVersion", m.PgVersion).
		Msg("Agent registered.")

	agent.UpdateDatabases(m.Databases)

	agent.ServeDbTask(s)

	return nil
}
