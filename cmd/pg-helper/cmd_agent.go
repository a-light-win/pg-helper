package main

import "github.com/a-light-win/pg-helper/internal/config"

type AgentCmd struct {
	config.AgentConfig
}

func (a *AgentCmd) Run(ctx *Context) error {
	// TODO: Implement the AgentCmd
	return nil
}
