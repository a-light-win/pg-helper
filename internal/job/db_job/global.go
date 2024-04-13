package db_job

import (
	"context"

	"github.com/a-light-win/pg-helper/internal/config"
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
