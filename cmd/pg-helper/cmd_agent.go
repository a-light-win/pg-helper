package main

import (
	"github.com/a-light-win/pg-helper/internal/agent"
	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/rs/zerolog/log"
)

type AgentCmd struct {
	config.AgentConfig
}

func (a *AgentCmd) Run(ctx *Context) error {
	agent_ := agent.New(&a.AgentConfig)

	utils.PrintCurrentLogLevel()
	log.Log().Msgf("pg-helper agent %s is start up", Version)

	if err := agent_.Init(); err != nil {
		return err
	}

	agent_.Run()
	return nil
}
