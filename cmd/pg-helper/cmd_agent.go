package main

import (
	"github.com/a-light-win/pg-helper/internal/agent"
	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/a-light-win/pg-helper/pkg/validate"
	"github.com/rs/zerolog/log"
)

type AgentCmd struct {
	config.AgentConfig
}

func (a *AgentCmd) Run() error {
	validator := validate.New()
	if err := validator.Struct(a.AgentConfig); err != nil {
		log.Error().Err(err).Msg("config validation failed")
		return err
	}

	utils.PrintCurrentLogLevel()
	log.Log().Msgf("pg-helper agent %s is start up", Version)

	agent_ := agent.New(&a.AgentConfig)
	if err := agent_.Init(); err != nil {
		return err
	}

	agent_.Run()
	return nil
}

func (a *AgentCmd) AfterApply() error {
	return a.AgentConfig.Db.AfterApply()
}
