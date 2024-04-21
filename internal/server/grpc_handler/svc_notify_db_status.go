package grpc_handler

import (
	"context"
	"errors"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (h *DbTaskSvcHandler) NotifyDbStatus(ctx context.Context, db *proto.Database) (*emptypb.Empty, error) {
	authInfo, err := authInfoWithType(ctx, AgentClient)
	if err != nil {
		return nil, err
	}

	agent := gd_.GetAgent(authInfo.ClientId)
	if agent == nil {
		err := errors.New("agent not found")
		log.Warn().Err(err).Str("AgentId", authInfo.ClientId).Msg("")
		return nil, err
	}

	agent.UpdateDatabase(db)
	return &emptypb.Empty{}, nil
}
