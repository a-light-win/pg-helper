package grpc_agent

import (
	"context"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/rs/zerolog/log"
)

type DbStatusSender struct {
	grpcClient proto.DbTaskSvcClient
	connCtx    context.Context
}

func (s *DbStatusSender) Handle(msg handler.NamedElement) error {
	log.Debug().Str("Dbname", msg.GetName()).Msg("Notify server the db status changed ...")
	_, err := s.grpcClient.NotifyDbStatus(s.connCtx, msg.(*proto.Database))
	log.Debug().Err(err).Str("Dbname", msg.GetName()).Msg("Notify server the db status changed finish")
	return err
}

func (s *DbStatusSender) Init(setter handler.GlobalSetter) error {
	return nil
}

func (s *DbStatusSender) PostInit(getter handler.GlobalGetter) error {
	s.grpcClient = getter.Get(constants.AgentKeyGrpcClient).(proto.DbTaskSvcClient)
	s.connCtx = getter.Get(constants.AgentKeyConnCtx).(context.Context)

	return nil
}
