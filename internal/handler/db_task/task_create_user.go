package db_task

import (
	"fmt"

	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/a-light-win/pg-helper/pkg/utils/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (h *DbTaskHandler) CreateUser(task *DbTask) (err error) {
	return h.DbApi.Query(func(q *db.Queries) error {
		return h.createUser(task, q)
	})
}

func (h *DbTaskHandler) createUser(task *DbTask, q *db.Queries) (err error) {
	connCtx := h.DbApi.ConnCtx
	log := log.With().
		Str("DbName", task.DbName).
		Str("Owner", task.Data.Owner).
		Logger()

	task.Status = db.DbTaskStatusRunning
	h.DbApi.UpdateTaskStatus(task.DbTask, nil)
	defer func() { setFinalTaskStatus(h.DbApi, task, err) }()

	db_, err := h.DbApi.GetDb(task.DbID, nil)
	if err != nil {
		log.Error().Err(err).
			Msg("Can not create user due to db error")
		return logger.NewAlreadyLoggedError(err, zerolog.ErrorLevel)
	}

	if !db_.IsNotExist() && db_.Stage != proto.DbStage_CreateUser {
		err = fmt.Errorf("db stage is not matching")
		log.Error().Err(err).
			Str("Stage", db_.Stage.String()).
			Str("Status", db_.Status.String()).
			Msg("Can not create user")
		return logger.NewAlreadyLoggedError(err, zerolog.ErrorLevel)
	}

	db_.Stage = proto.DbStage_CreateUser
	db_.Status = proto.DbStatus_Processing
	h.DbApi.UpdateDbStatus(db_, q)

	defer func() { setFinalDbStatus(h.DbApi, db_, err) }()

	exists, err := q.IsUserExists(connCtx, pgtype.Text{String: task.Data.Owner, Valid: true})
	if err != nil && err != pgx.ErrNoRows {
		log.Warn().Err(err).
			Str("Owner", task.Data.Owner).
			Str("DbName", task.DbName).
			Msg("Failed to check if user exists")
		return logger.NewAlreadyLoggedError(err, zerolog.WarnLevel)
	}

	if exists {
		log.Debug().Msg("User already exists")
		return nil
	}

	// Create User here
	conn := q.Conn()
	pgconn := conn.PgConn()
	escapedPassword, err := pgconn.EscapeString(task.Data.Password)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to escape password")
		return logger.NewAlreadyLoggedError(err, zerolog.WarnLevel)
	}

	_, err = conn.Exec(connCtx, fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'",
		task.Data.Owner,
		escapedPassword))
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create user")
		return err
	}

	log.Log().Msg("User created successfully")
	return nil
}
