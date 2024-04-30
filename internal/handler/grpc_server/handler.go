package grpc_server

import (
	"context"

	"github.com/a-light-win/pg-helper/api/proto"
	config "github.com/a-light-win/pg-helper/internal/config/server"
)

type DbTaskSvcHandler struct {
	proto.UnimplementedDbTaskSvcServer

	AgentDatas

	GrpcConfig *config.GrpcConfig

	QuitCtx context.Context
}

func NewDbTaskSvcHandler(config *config.GrpcConfig, quitCtx context.Context) *DbTaskSvcHandler {
	return &DbTaskSvcHandler{GrpcConfig: config, QuitCtx: quitCtx}
}
