package grpc_server

import (
	"errors"

	"github.com/a-light-win/pg-helper/api/proto"
	grpcAuth "github.com/a-light-win/pg-helper/pkg/auth/grpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *DbTaskSvcHandler) Register(m *proto.RegisterInstance, s proto.DbTaskSvc_RegisterServer) error {
	authInfo, ok := grpcAuth.LoadAuthInfo(s.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "no auth info")
	}
	logger := log.With().Str("AuthUuid", authInfo.Uuid).
		Str("AuthSubject", authInfo.Subject).
		Str("DbInstance", m.Name).
		Int32("PgVersion", m.PgVersion).
		Logger()

	if !authInfo.ValidateScope("agent") {
		err := errors.New("scope not allowed")
		logger.Warn().Err(err).Str("Scope", "agent").Msg("")
		return status.Error(codes.PermissionDenied, err.Error())
	}

	resouce := "dbInstance:" + m.Name
	if !authInfo.ValidateResource(resouce) {
		err := errors.New("resource not allowed")
		logger.Warn().Err(err).Str("Resource", resouce).Msg("")
		return status.Error(codes.PermissionDenied, err.Error())
	}

	instance := h.NewInstance(m.Name, m.PgVersion, &logger)
	logger.Log().Msg("Instance registered.")

	instance.UpdateDatabases(h.QuitCtx, m.Databases)

	instance.ServeDbTask(s)

	return nil
}
