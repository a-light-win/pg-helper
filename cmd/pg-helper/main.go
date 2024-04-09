package main

import (
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/alecthomas/kong"
	kongyaml "github.com/alecthomas/kong-yaml"
)

type Context struct{}

var Cli struct {
	LogLevel utils.LogLevel  `enum:"debug,info,warn,error,fatal" help:"Set the log level" default:"info"`
	Config   kong.ConfigFlag `help:"Load configuration from a file"`

	Version VersionCmd `cmd:"" help:"Print the version of pg-helper"`
	Agent   AgentCmd   `cmd:"" help:"Run the backup, restore or other pg commands in the background"`
	Serve   ServeCmd   `cmd:"" help:"The coordinator to manage the pg-helper agents"`
}

func main() {
	ctx := kong.Parse(&Cli, kong.Configuration(kongyaml.Loader, "/etc/pg-helper/config.yaml"))
	ctx.Run(&Context{})
}
