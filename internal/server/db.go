package server

import (
	"context"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
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
	for {
		m, err := migrate.New(
			s.Config.Db.MigrationsPath,
			s.Config.Db.Url(""))
		if err != nil {
			log.Warn().Err(err).Msg("Setup Migrate environment failed")
			time.Sleep(5 * time.Second)
			continue
		}

		if err := m.Up(); err != nil {
			if err != migrate.ErrNoChange {
				log.Error().Err(err).Msg("Migrate database failed")
				return err
			}
		}
		break
	}
	log.Log().Msg("Migrate database success")
	return nil
}
