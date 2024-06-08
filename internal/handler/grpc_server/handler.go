package grpc_server

import (
	"context"

	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/pkg/proto"
)

type DbJobSvcHandler struct {
	proto.UnimplementedDbJobSvcServer
	*DbInstanceManager

	GrpcConfig *config.GrpcConfig

	QuitCtx context.Context
}

func NewDbJobSvcHandler(config *config.GrpcConfig, quitCtx context.Context) *DbJobSvcHandler {
	return &DbJobSvcHandler{
		GrpcConfig: config, QuitCtx: quitCtx,
		DbInstanceManager: NewDbInstanceManager(),
	}
}
