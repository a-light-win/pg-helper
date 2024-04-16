package grpc_handler

import (
	"context"

	"github.com/a-light-win/pg-helper/api/proto"
)

type HandlerData struct {
	AgentDatas

	QuitCtx context.Context
}

var gd_ *HandlerData

func InitGlobalData() {
	gd_ = &HandlerData{}
}

type DbTaskSvcHandler struct {
	proto.UnimplementedDbTaskSvcServer
}
