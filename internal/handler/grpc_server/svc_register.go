package grpc_server

import (
	"github.com/a-light-win/pg-helper/api/proto"
	grpcAuth "github.com/a-light-win/pg-helper/pkg/auth/grpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *DbTaskSvcHandler) Register(m *proto.RegisterAgent, s proto.DbTaskSvc_RegisterServer) error {
	authInfo, ok := grpcAuth.LoadAuthInfo(s.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "no auth info")
	}
	if !authInfo.ValidateScope("agent") {
		return status.Error(codes.PermissionDenied, "no agent scope")
	}

	agent := h.NewAgent(authInfo.Subject, m.PgVersion)
	log.Log().
		Str("AgentId", authInfo.Subject).
		Int32("PgVersion", m.PgVersion).
		Msg("Agent registered.")

	agent.UpdateDatabases(h.QuitCtx, m.Databases)

	agent.ServeDbTask(s)

	return nil
}
