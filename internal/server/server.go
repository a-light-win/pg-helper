package server

import (
	"context"
	"time"

	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type Server struct {
	Config *config.Config
	Router *gin.Engine
	Pool   *pgxpool.Pool
}

func New(config *config.Config) *Server {
	r := gin.Default()

	server := Server{Config: config, Router: r}
	return &server
}

func (s *Server) Init() error {
	err := s.initWebServerByConfig()
	if err != nil {
		return err
	}
	err = s.registerRoutes()
	if err != nil {
		return err
	}
	err = s.initConnPool()
	if err != nil {
		return err
	}
	return s.migrateDb()
}

func (s *Server) initWebServerByConfig() error {
	s.Router.UseH2C = s.Config.Web.UseH2C
	s.Router.SetTrustedProxies(s.Config.Web.TrustedProxies)
	return nil
}

func (s *Server) registerRoutes() error {
	// TODO
	// s.Router.Post("/api/v1/db/create", handler.CreateDb)
	return nil
}

func (s *Server) initConnPool() error {
	// Initialize the database connection pool.
	poolConfig := s.Config.Db.NewPoolConfig()
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create a pool")
		return err
	}
	s.Pool = pool
	return nil
}

func (s *Server) migrateDb() error {
	log.Log().Msg("Start to migrate database")
	for {
		m, err := migrate.New(
			s.Config.Db.MigrationsPath,
			s.Config.Db.Url())
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

func (s *Server) Run() {
	log.Log().Msg("Start the web server")
	// TODO: Add graceful shutdown
	// TODO: customize the address and port
	err := s.Router.Run() // listen and serve on 0.0.0.0:8080
	if err != nil {
		log.Fatal().Err(err).Msg("Web server exit with error")
	}
}
