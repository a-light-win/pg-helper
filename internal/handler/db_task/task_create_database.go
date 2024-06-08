package db_task

import (
	"errors"
	"fmt"

	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/a-light-win/pg-helper/pkg/utils/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (h *DbTaskHandler) CreateDatabase(task *DbTask) error {
	return h.DbApi.Query(func(q *db.Queries) error {
		return h.createDatabase(task, q)
	})
}

func (h *DbTaskHandler) createDatabase(task *DbTask, q *db.Queries) (err error) {
	log := log.With().
		Str("DbName", task.DbName).
		Str("Action", string(task.Action)).
		Logger()

	database, err := h.DbApi.GetDbByName(task.DbName, q)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check if database exists")
		return logger.NewAlreadyLoggedError(err, zerolog.ErrorLevel)
	} else {
		if database.Owner != task.Data.Owner {
			err := errors.New("database exists with another owner")
			log.Error().Err(err).
				Str("actualOwner", database.Owner).
				Msg("")
			return logger.NewAlreadyLoggedError(err, zerolog.ErrorLevel)
		}
	}

	if !database.IsNotExist() && database.Stage != proto.DbStage_CreateUser && database.Stage != proto.DbStage_CreateDatabase {
		err := fmt.Errorf("db stage is not matching")
		log.Error().Err(err).Msg("Can not create database")
		return nil
	}

	task.Status = db.DbTaskStatusRunning
	h.DbApi.UpdateTaskStatus(task.DbTask, q)
	defer func() { setFinalTaskStatus(h.DbApi, task, err) }()

	database.Stage = proto.DbStage_CreateDatabase
	database.Status = proto.DbStatus_Processing
	h.DbApi.UpdateDbStatus(database, q)

	defer func() { setFinalDbStatus(h.DbApi, database, err) }()

	// Create database
	conn := q.Conn()
	_, err = conn.Exec(h.DbApi.ConnCtx, fmt.Sprintf("CREATE DATABASE %s OWNER %s",
		task.DbName, task.Data.Owner))
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create database")
		return logger.NewAlreadyLoggedError(err, zerolog.WarnLevel)
	}

	log.Log().Msg("Database created successfully")
	return nil
}
