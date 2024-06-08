package db

import (
	"context"
	"time"

	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/a-light-win/pg-helper/pkg/utils/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type DbApi struct {
	DbConfig *config.DbConfig
	DbPool   *pgxpool.Pool

	ConnCtx context.Context
	Cancel  context.CancelFunc

	DbStatusNotifier server.Producer
}

func (q *Queries) Conn() *pgx.Conn {
	switch q.db.(type) {
	case *pgxpool.Conn:
		return q.db.(*pgxpool.Conn).Conn()
	case *pgx.Conn:
		return q.db.(*pgx.Conn)
	default:
		return nil
	}
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

type (
	QueryFunc             func(*Queries) error
	QueryWithRollbackFunc func(pgx.Tx) error
)

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

func (api *DbApi) QueryWithRollback(queryFunc QueryWithRollbackFunc) error {
	conn, err := api.DbPool.Acquire(api.ConnCtx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire connection")
		return err
	}
	defer conn.Release()
	tx, err := conn.Begin(api.ConnCtx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to begin transaction")
		return err
	}
	defer tx.Rollback(api.ConnCtx)

	return queryFunc(tx)
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
		ID:          db.ID,
		Stage:       db.Stage,
		Status:      db.Status,
		UpdatedAt:   db.UpdatedAt,
		ExpiredAt:   db.ExpiredAt,
		LastJobID:   db.LastJobID,
		MigrateTo:   db.MigrateTo,
		MigrateFrom: db.MigrateFrom,
		ErrorMsg:    db.ErrorMsg,
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

	db_ := db.ToProto()
	db_.InstanceName = api.DbConfig.InstanceName
	api.DbStatusNotifier.Send(db_)
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
		if err != nil {
			if _, ok := err.(*logger.AlreadyLoggedError); !ok {
				log.Error().Err(err).
					Str("DbName", params.Name).
					Msg("Create db recored failed")
				return nil, logger.NewAlreadyLoggedError(err, zerolog.ErrorLevel)
			}
			return nil, err
		}
		return db, nil
	}

	db, err := q.CreateDb(api.ConnCtx, *params)
	if err != nil {
		log.Error().Err(err).
			Str("DbName", params.Name).
			Msg("Create db recored failed")
		return nil, logger.NewAlreadyLoggedError(err, zerolog.ErrorLevel)
	}
	return &db, nil
}

func (api *DbApi) CreateDbTask(params *CreateDbTaskParams, q *Queries) (*DbTask, error) {
	if q == nil {
		var task *DbTask
		var err error
		err = api.Query(func(q *Queries) error {
			task, err = api.CreateDbTask(params, q)
			return err
		})
		if err != nil {
			if _, ok := err.(*logger.AlreadyLoggedError); !ok {
				log.Error().Err(err).
					Str("DbName", params.DbName).
					Str("Action", string(params.Action)).
					Msg("Create db_task recored failed")
				return nil, logger.NewAlreadyLoggedError(err, zerolog.ErrorLevel)
			}
			return nil, err
		}
		return task, nil
	}

	task, err := q.CreateDbTask(api.ConnCtx, *params)
	if err != nil {
		log.Error().Err(err).
			Str("DbName", params.DbName).
			Str("Action", string(params.Action)).
			Msg("Create db_task recored failed")
		return nil, logger.NewAlreadyLoggedError(err, zerolog.ErrorLevel)
	}
	return &task, nil
}

func (api *DbApi) ListDbs(q *Queries) ([]Db, error) {
	if q == nil {
		var dbs []Db
		var err error
		api.Query(func(q *Queries) error {
			dbs, err = api.ListDbs(q)
			return err
		})
		return dbs, err
	}

	dbs, err := q.ListDbs(api.ConnCtx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return []Db{}, nil
		}
		return nil, err
	}
	return dbs, nil
}

func (api *DbApi) ToProtoDatabases(dbs []Db) []*proto.Database {
	if len(dbs) == 0 {
		return []*proto.Database{}
	}

	databases := make([]*proto.Database, len(dbs))
	for i := range dbs {
		databases[i] = dbs[i].ToProto()
	}
	return databases
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

	if err := MigrateUp(db_); err != nil {
		log.Error().Err(err).Msg("Migrate database failed")
		return err
	}

	log.Log().Msg("Migrate database success")
	return nil
}
