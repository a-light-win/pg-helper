package grpc_server

import (
	"context"
	"errors"

	grpcAuth "github.com/a-light-win/pg-helper/pkg/auth/grpc"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *DbJobSvcHandler) NotifyDbStatus(ctx context.Context, db *proto.Database) (*emptypb.Empty, error) {
	authInfo, ok := grpcAuth.LoadAuthInfo(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "no auth info")
	}
	if !authInfo.ValidateScope("agent") {
		return nil, status.Error(codes.PermissionDenied, "no scope permission")
	}
	if !authInfo.ValidateResource("dbInstance:" + db.InstanceName) {
		return nil, status.Error(codes.PermissionDenied, "no resource permission")
	}

	instance := h.GetInstance(db.InstanceName)
	if instance == nil {
		err := errors.New("db instance not found")
		log.Warn().Err(err).Str("InstanceName", db.InstanceName).Msg("")
		return nil, err
	}

	instance.UpdateDatabase(db)
	return &emptypb.Empty{}, nil
}
