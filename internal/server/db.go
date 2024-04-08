package server

import (
	"context"
	"time"

	"github.com/a-light-win/pg-helper/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/rs/zerolog/log"
)

func (s *Server) initDb() error {
	err := s.initConnPool()
	if err != nil {
		return err
	}
	return s.migrateDb()
}

func (s *Server) initConnPool() error {
	// Initialize the database connection pool.
	poolConfig := s.Config.Db.NewPoolConfig()
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create a pool")
		return err
	}
	s.DbPool = pool
	return nil
}

func (s *Server) migrateDb() error {
	log.Log().Msg("Start to migrate database")

	db_ := stdlib.OpenDBFromPool(s.DbPool)
	defer db_.Close()

	// Ensure the database connection is ready.
	for {
		if err := db_.Ping(); err != nil {
			log.Warn().Err(err).Msg("Ping database failed")
			time.Sleep(5 * time.Second)
			continue
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
