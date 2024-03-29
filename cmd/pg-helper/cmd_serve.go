package main

import (
	config_ "github.com/a-light-win/pg-helper/internal/config"
	server_ "github.com/a-light-win/pg-helper/internal/server"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd)

	viper.SetDefault("web::use_h2c", true)
	viper.SetDefault("web::trusted_proxies", []string{})

	viper.SetDefault("db::host", "127.0.0.1")
	viper.SetDefault("db::port", 5432)
	viper.SetDefault("db::user", "postgres")
	viper.SetDefault("db::db_name", "postgres")
	viper.SetDefault("db::reserve_db_names", []string{"postgres"})
	viper.SetDefault("db::max_conns", 4)
	viper.SetDefault("db::migrations_path", "file:///usr/share/pg-helper/migrations")
	viper.BindEnv("db::password_file", "PG_HELPER_DB_PASSWORD_FILE")
	viper.BindEnv("db::password", "PG_HELPER_DB_PASSWORD")
	viper.BindEnv("db::host", "PG_HELPER_DB_HOST")
}

var (
	config   = config_.Config{}
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		Run:   serve,
	}
)

func serve(cmd *cobra.Command, args []string) {
	gin.SetMode(gin.ReleaseMode)

	if err := viper.ReadInConfig(); err == nil {
		log.Log().Str("file", viper.ConfigFileUsed()).Msg("Load config")
	} else {
		log.Warn().Err(err).Msg("Failed to load config")
	}

	viper.Unmarshal(&config)
	server := server_.New(&config)

	server.InitLogger()
	log.Log().Msgf("pg-helper %s is start up", Version)

	if server.Init() != nil {
		return
	}

	server.Run()
}
