package main

import (
	config_ "github.com/a-light-win/pg-helper/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd)

	viper.SetDefault("web::use_h2c", true)
	viper.SetDefault("web::trusted_proxies", []string{})
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

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	err := r.Run() // listen and serve on 0.0.0.0:8080
	if err != nil {
		// TODO: log error and exit
		panic(err)
		// os.exit(1)
	}
}
