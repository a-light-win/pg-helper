package grpc_server

import (
	"context"

	"github.com/a-light-win/pg-helper/api/proto"
	config "github.com/a-light-win/pg-helper/internal/config/server"
)

type HandlerData struct {
	AgentDatas
	GrpcConfig *config.GrpcConfig

	QuitCtx context.Context
}

var gd_ *HandlerData

func InitGlobalData(config *config.GrpcConfig, quitCtx context.Context) {
	gd_ = &HandlerData{GrpcConfig: config, QuitCtx: quitCtx}
}

type DbTaskSvcHandler struct {
	proto.UnimplementedDbTaskSvcServer
}
