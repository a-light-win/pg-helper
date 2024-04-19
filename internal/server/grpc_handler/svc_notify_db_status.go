package grpc_handler

import (
	"context"
	"errors"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/rs/zerolog/log"
)

func (h *DbTaskSvcHandler) NotifyDbStatus(ctx context.Context, db *proto.Database) error {
	// TODO: set the AgentId in the context
	authInfo, err := authInfoWithType(ctx, AgentClient)
	if err != nil {
		return err
	}

	agent := gd_.GetAgent(authInfo.ClientId)
	if agent == nil {
		err := errors.New("agent not found")
		log.Warn().Err(err).Str("AgentId", authInfo.ClientId).Msg("")
		return err
	}

	agent.UpdateDatabase(db)
	return nil
}
