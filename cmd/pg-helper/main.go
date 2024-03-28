package main

import (
	"fmt"
	"os"
	"path/filepath"

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
		home, err := os.UserCacheDir()
		cobra.CheckErr(err)

		home_config_path := filepath.Join(home, ".config", "pg-helper")
		viper.AddConfigPath(home_config_path)
		viper.AddConfigPath("/etc/pg-helper/")
		viper.SetConfigName("config.yaml")
	}

	viper.SetEnvPrefix("PG_HELPER")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
