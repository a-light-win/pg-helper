package grpc_server

import (
	"context"

	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/pkg/proto"
)

type DbTaskSvcHandler struct {
	proto.UnimplementedDbTaskSvcServer
	*DbInstanceManager

	GrpcConfig *config.GrpcConfig

	QuitCtx context.Context
}

func NewDbTaskSvcHandler(config *config.GrpcConfig, quitCtx context.Context) *DbTaskSvcHandler {
	return &DbTaskSvcHandler{
		GrpcConfig: config, QuitCtx: quitCtx,
		DbInstanceManager: NewDbInstanceManager(),
	}
}
