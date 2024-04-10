package agent

import (
	"context"
	"time"

	"github.com/a-light-win/pg-helper/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/rs/zerolog/log"
)

func (a *Agent) initDb() error {
	err := a.initConnPool()
	if err != nil {
		return err
	}
	return a.migrateDb()
}

func (a *Agent) initConnPool() error {
	// Initialize the database connection pool.
	poolConfig := a.Config.Db.NewPoolConfig()
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create a pool")
		return err
	}
	a.DbPool = pool
	return nil
}

func (a *Agent) migrateDb() error {
	log.Log().Msg("Start to migrate database")

	db_ := stdlib.OpenDBFromPool(a.DbPool)
	defer db_.Close()

	// Ensure the database connection is ready.
	for {
		if err := db_.Ping(); err != nil {
			log.Warn().Err(err).Msg("Ping database failed")

			select {
			case <-a.QuitCtx.Done():
				log.Log().Err(a.QuitCtx.Err()).Msg("Receive quit signal when ping database")
				return a.QuitCtx.Err()
			case <-time.After(5 * time.Second):
				continue
			}
		}
		break
	}

	if err := db.MigrateUp(db_); err != nil {
		log.Error().Err(err).Msg("Migrate database failed")
		return err
	}

	log.Log().Msg("Migrate database success")
	return nil
}