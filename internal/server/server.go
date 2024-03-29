package server

import (
	"context"
	"log"

	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	Config *config.Config
	Router *gin.Engine
	Pool   *pgxpool.Pool
}

func New(config *config.Config) *Server {
	r := gin.Default()

	// Initialize the database connection pool.
	poolConfig := config.Db.NewPoolConfig()
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal("Failed to create a pool, error: ", err)
	}

	server := Server{config, r, pool}
	return &server
}

func (s *Server) Init() {
	s.initWebServerByConfig()
	s.registerRoutes()
	s.migrateDb()
}

func (s *Server) initWebServerByConfig() {
	s.Router.UseH2C = s.Config.Web.UseH2C
	s.Router.SetTrustedProxies(s.Config.Web.TrustedProxies)
}

func (s *Server) registerRoutes() {
	// TODO
	// s.Router.Post("/api/v1/db/create", handler.CreateDb)
}

func (s *Server) migrateDb() {
	for {
		m, err := migrate.New(
			s.Config.Db.MigrationsPath,
			s.Config.Db.Url())
		if err != nil {
			log.Fatal(err)
			continue
		}

		if err := m.Up(); err != nil {
			if err != migrate.ErrNoChange {
				log.Fatal(err)
			}
		}
		break
	}
}

func (s *Server) Run() {
	err := s.Router.Run() // listen and serve on 0.0.0.0:8080
	if err != nil {
		// TODO: log error and exit
		panic(err)
		// os.exit(1)
	}
}
