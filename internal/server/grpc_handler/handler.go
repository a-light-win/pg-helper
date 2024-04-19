package grpc_handler

import (
	"context"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/config"
)

type HandlerData struct {
	AgentDatas
	GrpcConfig *config.GrpcServerConfig

	QuitCtx context.Context
}

var gd_ *HandlerData

func InitGlobalData() {
	gd_ = &HandlerData{}
}

type DbTaskSvcHandler struct {
	proto.UnimplementedDbTaskSvcServer
}
