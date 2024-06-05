package grpc_agent

import (
	"errors"

	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/a-light-win/pg-helper/pkg/utils/logger"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func (h *GrpcAgentHandler) migrateOutDatabase(task *proto.DbTask) error {
	taskData := task.GetMigrateOutDatabase()

	return h.DbApi.QueryWithRollback(func(q *db.Queries) error {
		return migrateOutDatabase(h.DbApi, taskData, q)
	})
}

func migrateOutDatabase(dbApi *db.DbApi, taskData *proto.MigrateOutDatabaseTask, q *db.Queries) error {
	database, err := dbApi.GetDbByName(taskData.Name, q)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		log.Warn().Err(err).
			Str("Name", taskData.Name).
			Msg("Migrate database out failed")
		return logger.NewAlreadyLoggedError(err, zerolog.WarnLevel)
	}

	// TODO create db task and update the status
	//
	if database.Stage == proto.DbStage_Idle ||
		database.Stage == proto.DbStage_Dropping ||
		database.Stage == proto.DbStage_DropCompleted {
		log.Debug().
			Str("Name", taskData.Name).
			Msg("Database is already migrated out")
		return nil
	}

	if database.Stage != proto.DbStage_Ready {
		err := errors.New("database is not in ready stage, can not migrate out")
		log.Warn().Err(err).
			Str("Name", taskData.Name).
			Msg("Migrate database out failed")
		return err
	}

	database.Stage = proto.DbStage_Idle
	if taskData.ExpiredAt.IsValid() {
		database.ExpiredAt.Scan(taskData.ExpiredAt.AsTime())
	} else {
		// TODO: set to default expired time?
		// database.ExpiredAt(time.Now() + dbApi.DbConfig.DurationToDropIdleDb)
	}
	// TODO: Add cron task to drop the database

	return dbApi.UpdateDbStatus(database, q)
}
