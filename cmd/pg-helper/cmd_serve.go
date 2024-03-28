package main

import (
	"fmt"

	config_ "github.com/a-light-win/pg-helper/internal/config"
	"github.com/gin-gonic/gin"
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
	viper.BindEnv("db::password_file", "PG_HELPER_DB_PASSWORD_FILE")
	viper.BindEnv("db::password", "PG_HELPER_DB_PASSWORD")
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
	r := gin.Default()

	viper.Unmarshal(&config)
	r.UseH2C = config.Web.UseH2C
	r.SetTrustedProxies(config.Web.TrustedProxies)

	fmt.Println("The config is")
	fmt.Println(config.Db.PasswordFile)
	fmt.Println(config.Db.Host)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// handler = handler_.Handler{Db: db_.New()}

	err := r.Run() // listen and serve on 0.0.0.0:8080
	if err != nil {
		// TODO: log error and exit
		panic(err)
		// os.exit(1)
	}
}
