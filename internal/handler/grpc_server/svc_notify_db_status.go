package grpc_server

import (
	"context"
	"errors"

	"github.com/a-light-win/pg-helper/api/proto"
	grpcAuth "github.com/a-light-win/pg-helper/pkg/auth/grpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *DbTaskSvcHandler) NotifyDbStatus(ctx context.Context, db *proto.Database) (*emptypb.Empty, error) {
	authInfo, ok := grpcAuth.LoadAuthInfo(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no auth info")
	}
	if !authInfo.ValidateScope("agent") {
		return nil, status.Error(codes.PermissionDenied, "no agent scope")
	}

	agent := h.GetAgent(authInfo.Subject)
	if agent == nil {
		err := errors.New("agent not found")
		log.Warn().Err(err).Str("AgentId", authInfo.Subject).Msg("")
		return nil, err
	}

	agent.UpdateDatabase(h.QuitCtx, db)
	return &emptypb.Empty{}, nil
}
