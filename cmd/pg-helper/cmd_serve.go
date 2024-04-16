package main

import (
	config_ "github.com/a-light-win/pg-helper/internal/config"
	server_ "github.com/a-light-win/pg-helper/internal/server"
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type ServeCmd struct {
	config_.ServerConfig
}

func (s *ServeCmd) Run(ctx *Context) error {
	gin.SetMode(gin.ReleaseMode)

	server := server_.New(&s.ServerConfig)

	utils.PrintCurrentLogLevel()
	log.Log().Msgf("pg-helper %s is start up", Version)

	if err := server.Init(); err != nil {
		return err
	}

	server.Run()
	return nil
}

/*
func init() {
	rootCmd.AddCommand(serveCmd)

	viper.SetDefault("web::use_h2c", true)
	viper.SetDefault("web::trusted_proxies", []string{})

	viper.SetDefault("db::host", "127.0.0.1")
	viper.SetDefault("db::port", 5432)
	viper.SetDefault("db::user", "postgres")
	viper.SetDefault("db::db_name", "postgres")
	viper.SetDefault("db::reserve_db_names", []string{"postgres", "template0", "template1"})
	viper.SetDefault("db::max_conns", 4)
	viper.SetDefault("db::migrations_path", "file:///usr/share/pg-helper/migrations")
	viper.SetDefault("db::backup_root_path", "/var/lib/pg-helper/backups")
	viper.SetDefault("remote_helper::url_template", "http://pg-%d")

	viper.BindEnv("db::password_file", "PG_HELPER_DB_PASSWORD_FILE")
	viper.BindEnv("db::password", "PG_HELPER_DB_PASSWORD")
	viper.BindEnv("db::host", "PG_HELPER_DB_HOST")
	viper.BindEnv("db::current_db_version", "PG_MAJOR")
}
*/
