package main

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	viper_ "github.com/spf13/viper"
)

var (
	viper   = viper_.NewWithOptions(viper_.KeyDelimiter("::"))
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "pg-helper",
		Short: "Helpe to manage multiple databases in a postgresql instance",
		Long: `pg-helper is a tool to help manage multiple databases in a postgresql instance.
It can be used to create a new database and its owner,
or migrade a database from an old pg version to current pg version.`,
	}
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pg-helper)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigFile("/etc/pg-helper/config.yaml")
	}

	viper.SetEnvPrefix("PG_HELPER")
	viper.AutomaticEnv()
}
