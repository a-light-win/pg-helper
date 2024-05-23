package grpc_agent

import (
	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/go-playground/validator/v10"
)

type GrpcAgentHandler struct {
	DbApi       *db.DbApi
	GrpcClient  proto.DbTaskSvcClient
	JobProducer server.Producer
	Validator   *validator.Validate
}

func NewGrpcAgentHandler(dbApi *db.DbApi, grpcClient proto.DbTaskSvcClient, jobProducer server.Producer) *GrpcAgentHandler {
	return &GrpcAgentHandler{
		DbApi:       dbApi,
		GrpcClient:  grpcClient,
		JobProducer: jobProducer,
		Validator:   NewValidator(),
	}
}

func (h *GrpcAgentHandler) handle(task *proto.DbTask) error {
	switch task.Task.(type) {
	case *proto.DbTask_CreateDatabase:
		return h.createDatabase(task)
	}
	return nil
}
