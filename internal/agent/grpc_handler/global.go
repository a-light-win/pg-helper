package grpc_handler

import (
	"context"

	"github.com/a-light-win/pg-helper/api/proto"
	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/jackc/pgx/v5/pgxpool"
)

type grpcGlobalData struct {
	DbPool   *pgxpool.Pool
	DbConfig *config.DbConfig

	GrpcClient proto.DbTaskSvcClient

	JobProducer *job.JobProducer

	ConnCtx context.Context
}

var gd_ *grpcGlobalData

func InitGlobalData(dbPool *pgxpool.Pool, dbConfig *config.DbConfig, grpcClient proto.DbTaskSvcClient, jobProducer *job.JobProducer, connCtx context.Context) {
	gd_ = &grpcGlobalData{
		DbPool:      dbPool,
		DbConfig:    dbConfig,
		GrpcClient:  grpcClient,
		JobProducer: jobProducer,
		ConnCtx:     connCtx,
	}
}
