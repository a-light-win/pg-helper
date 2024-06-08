package grpc_agent

import (
	"context"

	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/a-light-win/pg-helper/pkg/server"
)

type DbStatusSender struct {
	grpcClient proto.DbJobSvcClient
	connCtx    context.Context
}

func (s *DbStatusSender) Handle(msg server.NamedElement) error {
	_, err := s.grpcClient.NotifyDbStatus(s.connCtx, msg.(*proto.Database))
	return err
}

func (s *DbStatusSender) Init(setter server.GlobalSetter) error {
	return nil
}

func (s *DbStatusSender) PostInit(getter server.GlobalGetter) error {
	s.grpcClient = getter.Get(constants.AgentKeyGrpcClient).(proto.DbJobSvcClient)
	s.connCtx = getter.Get(constants.AgentKeyConnCtx).(context.Context)

	return nil
}
