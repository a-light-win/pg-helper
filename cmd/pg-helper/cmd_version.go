package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version   = "0.0.1"
	GoVersion = "1.22"
	buildInfo bool
)

func init() {
	rootCmd.AddCommand(versionCmd)

	versionCmd.Flags().BoolVarP(&buildInfo, "build-info", "b", false, "Print build information")
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of pg-helper",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pg-helper", Version)
		if buildInfo {
			fmt.Println("Built by: go", GoVersion)
		}
	},
}
