package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(migrateCmd)
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate a database from an old pg version to current pg version",
	Run:   migrate,
}

func migrate(cmd *cobra.Command, args []string) {
	// TODO: Implement the migrate command
	fmt.Println("migrate")
}
