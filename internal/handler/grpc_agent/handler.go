package grpc_agent

import (
	"context"

	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/go-playground/validator/v10"
)

type GrpcAgentHandler struct {
	DbApi *db.DbApi

	GrpcClient proto.DbJobSvcClient
	QuitCtx    context.Context

	JobProducer server.Producer
	Validator   *validator.Validate
}

func NewGrpcAgentHandler(dbApi *db.DbApi, grpcClient proto.DbJobSvcClient, jobProducer server.Producer, quitCtx context.Context) *GrpcAgentHandler {
	return &GrpcAgentHandler{
		DbApi:       dbApi,
		GrpcClient:  grpcClient,
		JobProducer: jobProducer,
		QuitCtx:     quitCtx,
		Validator:   NewValidator(),
	}
}

func (h *GrpcAgentHandler) handle(task *proto.DbJob) error {
	switch task.Job.(type) {
	case *proto.DbJob_CreateDatabase:
		request := NewCreateDatabaseRequest(task)
		return request.Process(h)
	case *proto.DbJob_MigrateOutDatabase:
		request := NewMigrateOutDatabaseRequest(task)
		return request.Process(h)
	}
	return nil
}
