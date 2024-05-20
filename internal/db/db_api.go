package db

import (
	"context"
	"time"

	"github.com/a-light-win/pg-helper/api/proto"
	migrate "github.com/a-light-win/pg-helper/db"
	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type DbApi struct {
	DbConfig *config.DbConfig
	DbPool   *pgxpool.Pool

	ConnCtx context.Context
	Cancel  context.CancelFunc
}

func NewDbApi(config *config.DbConfig) (*DbApi, error) {
	pool, err := initConnPool(config)
	if err != nil {
		return nil, err
	}

	connCtx, cancel := context.WithCancel(context.Background())
	return &DbApi{
		DbConfig: config,
		DbPool:   pool,
		ConnCtx:  connCtx,
		Cancel:   cancel,
	}, nil
}

func initConnPool(dbConfig *config.DbConfig) (*pgxpool.Pool, error) {
	// Initialize the database connection pool.
	poolConfig := dbConfig.NewPoolConfig()
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create a pool")
		return nil, err
	}
	return pool, nil
}

type QueryFunc func(*Queries) error

func (api *DbApi) Acquire() (*pgxpool.Conn, error) {
	return api.DbPool.Acquire(api.ConnCtx)
}

func (api *DbApi) Query(queryFunc QueryFunc) error {
	conn, err := api.DbPool.Acquire(api.ConnCtx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire connection")
		return err
	}
	defer conn.Release()
	q := New(conn)

	return queryFunc(q)
}

func (api *DbApi) UpdateTaskStatus(taskId uuid.UUID, status DbTaskStatus, reason string) error {
	dbTaskParams := SetDbTaskStatusParams{ID: taskId, Status: status, ErrReason: reason}
	return api.Query(func(q *Queries) error {
		return q.SetDbTaskStatus(api.ConnCtx, dbTaskParams)
	})
}

func (api *DbApi) UpdateDbStatus(dbId int64, stage proto.DbStage, status proto.DbStatus) error {
	dbStatusParams := SetDbStatusParams{ID: dbId, Stage: stage, Status: status}
	return api.Query(func(q *Queries) error {
		return q.SetDbStatus(api.ConnCtx, dbStatusParams)
	})
}

func (api *DbApi) MigrateDB(quitCtx context.Context) error {
	log.Log().Msg("Start to migrate database")

	db_ := stdlib.OpenDBFromPool(api.DbPool)
	defer db_.Close()

	// Ensure the database connection is ready.
	for {
		if err := db_.Ping(); err != nil {
			log.Warn().Err(err).Msg("Ping database failed")

			select {
			case <-quitCtx.Done():
				log.Log().Err(quitCtx.Err()).Msg("Receive quit signal when ping database")
				return quitCtx.Err()
			case <-time.After(5 * time.Second):
				continue
			}
		}
		break
	}

	if err := migrate.MigrateUp(db_); err != nil {
		log.Error().Err(err).Msg("Migrate database failed")
		return err
	}

	log.Log().Msg("Migrate database success")
	return nil
}
