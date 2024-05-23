package db

import (
	"context"
	"time"

	migrate "github.com/a-light-win/pg-helper/db"
	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

type DbApi struct {
	DbConfig *config.DbConfig
	DbPool   *pgxpool.Pool

	ConnCtx context.Context
	Cancel  context.CancelFunc

	DbStatusNotifier handler.Producer
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

func (api *DbApi) UpdateTaskStatus(task *DbTask, q *Queries) error {
	if q == nil {
		return api.Query(func(q *Queries) error {
			return api.UpdateTaskStatus(task, q)
		})
	}

	dbTaskParams := SetDbTaskStatusParams{
		ID:        task.ID,
		Status:    task.Status,
		ErrReason: task.Data.ErrReason,
		UpdatedAt: task.UpdatedAt,
	}

	newTask, err := q.SetDbTaskStatus(api.ConnCtx, dbTaskParams)
	if err != nil {
		log.Warn().Err(err).
			Interface("TaskID", task.ID).
			Str("Status", string(task.Status)).
			Time("UpdatedAt", task.UpdatedAt.Time).
			Msg("can not update task status")
		return err
	}

	task.UpdatedAt = newTask.UpdatedAt

	return nil
}

func (api *DbApi) UpdateDbStatus(db *Db, q *Queries) error {
	if q == nil {
		return api.Query(func(q *Queries) error {
			return api.UpdateDbStatus(db, q)
		})
	}

	dbStatusParams := SetDbStatusParams{
		ID:        db.ID,
		Stage:     db.Stage,
		Status:    db.Status,
		UpdatedAt: db.UpdatedAt,
	}

	newDb, err := q.SetDbStatus(api.ConnCtx, dbStatusParams)
	if err != nil {
		log.Warn().Err(err).
			Str("DbName", db.Name).
			Str("Stage", db.Stage.String()).
			Str("Status", db.Status.String()).
			Interface("UpdatedAt", db.UpdatedAt).
			Msg("Can not change db status to database")
		return err
	}

	db.UpdatedAt = newDb.UpdatedAt
	api.NotifyDbStatusChanged(db)

	return nil
}

func (api *DbApi) NotifyDbStatusChanged(db *Db) {
	log.Info().Str("DbName", db.Name).
		Str("Stage", db.Stage.String()).
		Str("Status", db.Status.String()).
		Interface("UpdatedAt", db.UpdatedAt).
		Msg("Database status changed")

	api.DbStatusNotifier.Send(db.ToProto())
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
