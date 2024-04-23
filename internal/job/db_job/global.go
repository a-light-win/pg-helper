package db_job

import (
	"context"

	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DbJobGlobalData struct {
	DbPool  *pgxpool.Pool
	ConnCtx context.Context

	DbConfig *config.DbConfig
}

var gd_ *DbJobGlobalData

func InitGlobalData(pool *pgxpool.Pool, connCtx context.Context, dbConfig *config.DbConfig) {
	gd_ = &DbJobGlobalData{
		DbPool:   pool,
		ConnCtx:  connCtx,
		DbConfig: dbConfig,
	}
}
