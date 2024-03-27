package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "0.0.1"

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of pg-helper",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("pg-helper v", Version)
	},
}
