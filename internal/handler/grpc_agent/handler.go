package grpc_agent

import (
	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/go-playground/validator/v10"
)

type GrpcAgentHandler struct {
	DbApi       *db.DbApi
	GrpcClient  proto.DbTaskSvcClient
	JobProducer *job.JobProducer
	Validator   *validator.Validate
}

func NewGrpcAgentHandler(dbApi *db.DbApi, grpcClient proto.DbTaskSvcClient, jobProducer *job.JobProducer) *GrpcAgentHandler {
	return &GrpcAgentHandler{
		DbApi:       dbApi,
		GrpcClient:  grpcClient,
		JobProducer: jobProducer,
		Validator:   NewValidator(),
	}
}

func (h *GrpcAgentHandler) Run(task *proto.DbTask) {
	if err := h.handle(task); err != nil {
		// TODO: log here?
	}
}

func (h *GrpcAgentHandler) handle(task *proto.DbTask) error {
	switch task.Task.(type) {
	case *proto.DbTask_CreateDatabase:
		return h.createDatabase(task)
	}
	return nil
}
