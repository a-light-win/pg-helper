package main

import (
	"os"

	"github.com/a-light-win/pg-helper/pkg/utils/logger"
	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
)

var Cli struct {
	LogLevel logger.LogLevel `enum:"debug,info,warn,error,fatal" help:"Set the log level" default:"info"`
	Config   kong.ConfigFlag `help:"Load configuration from a file"`

	Version VersionCmd `cmd:"" help:"Print the version of pg-helper"`
	Agent   AgentCmd   `cmd:"" help:"Run the backup, restore or other pg commands in the background"`
	Serve   ServeCmd   `cmd:"" help:"The coordinator to manage the pg-helper agents"`

	GenKey GenKeyCmd `cmd:"" help:"Generate a new Ed25519 key pair"`
	GenJwt GenJwtCmd `cmd:"" help:"Generate a new JWT token"`
}

func main() {
	ctx := kong.Parse(&Cli, kong.Configuration(kongyaml.Loader, "/etc/pg-helper/config.yaml"))
	err := ctx.Run()
	if err != nil {
		os.Exit(1)
	}
}
