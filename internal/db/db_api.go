package db

import (
	"context"
	"time"

	migrate "github.com/a-light-win/pg-helper/db"
	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (api *DbApi) UpdateDbStatus(db_ *Db, q *Queries) error {
	if q == nil {
		return api.Query(func(q *Queries) error {
			return api.UpdateDbStatus(db_, q)
		})
	}

	dbStatusParams := SetDbStatusParams{
		ID:        db_.ID,
		Stage:     db_.Stage,
		Status:    db_.Status,
		UpdatedAt: db_.UpdatedAt,
	}

	newDb, err := q.SetDbStatus(api.ConnCtx, dbStatusParams)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		return err
	}

	db_ = &newDb

	// TODO: notify the status change

	return nil
}

func (api *DbApi) GetDb(dbId int64, q *Queries) (*Db, error) {
	if q == nil {
		var db *Db
		var err error
		api.Query(func(q *Queries) error {
			db, err = api.GetDb(dbId, q)
			return err
		})
		return db, err
	}

	db, err := q.GetDbByID(api.ConnCtx, dbId)
	if err != nil {
		return nil, err
	}
	return &db, nil
}

func (api *DbApi) GetDbByName(name string, q *Queries) (*Db, error) {
	if q == nil {
		var db *Db
		var err error
		api.Query(func(q *Queries) error {
			db, err = api.GetDbByName(name, q)
			return err
		})
		return db, err
	}

	db, err := q.GetDbByName(api.ConnCtx, name)
	if err != nil {
		return nil, err
	}

	return &db, nil
}

func (api *DbApi) CreateDb(params *CreateDbParams, q *Queries) (*Db, error) {
	if q == nil {
		var db *Db
		var err error
		api.Query(func(q *Queries) error {
			db, err = api.CreateDb(params, q)
			return err
		})
		return db, err
	}

	db, err := q.CreateDb(api.ConnCtx, *params)
	if err != nil {
		return nil, err
	}
	return &db, err
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
