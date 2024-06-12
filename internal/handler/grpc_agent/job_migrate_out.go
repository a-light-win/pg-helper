package grpc_agent

import (
	"errors"

	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/handler/db_task"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/a-light-win/pg-helper/pkg/utils"
	"github.com/a-light-win/pg-helper/pkg/utils/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type MigrateOutDatabaseRequest struct {
	*proto.MigrateOutDatabaseJob
	JobId uuid.UUID
}

func NewMigrateOutDatabaseRequest(task *proto.DbJob) *MigrateOutDatabaseRequest {
	taskData := task.GetMigrateOutDatabase()
	return &MigrateOutDatabaseRequest{
		MigrateOutDatabaseJob: taskData,
		JobId:                 utils.StringToUuid(task.JobId),
	}
}

func (r *MigrateOutDatabaseRequest) Process(h *GrpcAgentHandler) error {
	return h.DbApi.QueryWithRollback(func(tx pgx.Tx) error {
		return r.process(h, tx)
	})
}

func (r *MigrateOutDatabaseRequest) process(h *GrpcAgentHandler, tx pgx.Tx) error {
	dbApi := h.DbApi
	q := db.New(tx)

	database, err := dbApi.GetDbByName(r.Name, q)
	if err != nil {
		if err != pgx.ErrNoRows {
			log.Error().Err(err).
				Str("Name", r.Name).
				Msg("Migrate database out failed")
			return logger.NewAlreadyLoggedError(err, zerolog.WarnLevel)
		}
	}

	if database.IsNotExist() {
		h.DbApi.NotifyDbStatusChanged(database)
		return nil
	}

	// if JobId exists and equal to the LastJobId,
	// loading the previous failed/cancelled tasks and retry
	// else new a JobId, and create new tasks.
	if r.JobId != uuid.Nil && r.JobId == database.LastJobID {
		// TODO: Load last failed job and retry it from the first failed task
		return nil
	}

	if database.IsAlreadyIdle() {
		log.Debug().
			Str("Name", r.Name).
			Msg("Database is already migrated out")
		h.DbApi.NotifyDbStatusChanged(database)
		return nil
	}

	if !database.IsReadyToUse() {
		err := errors.New("database is not in ready stage, can not migrate out")
		log.Warn().Err(err).
			Str("Name", r.Name).
			Msg("Migrate database out failed")
		h.DbApi.NotifyDbStatusChanged(database)
		return err
	}

	r.JobId = uuid.New()
	database.LastJobID = r.JobId
	database.Stage = proto.DbStage_Idle
	database.Status = proto.DbStatus_Processing
	database.MigrateTo = r.MigrateTo
	if r.ExpiredAt.IsValid() {
		database.ExpiredAt.Scan(r.ExpiredAt.AsTime())
	} else {
		// TODO: set to default expired time?
		// database.ExpiredAt(time.Now() + dbApi.DbConfig.DurationToDropIdleDb)
	}
	h.DbApi.UpdateDbStatus(database, q)

	// New a task to Idle the database
	dbTaskParams := db.CreateDbTaskParams{
		JobID:  r.JobId,
		DbID:   database.ID,
		DbName: database.Name,
		Action: db.DbActionMigrateOut,
		Reason: r.Reason,
		Status: db.DbTaskStatusPending,
		Data:   db.DbTaskData{},
	}

	migrateOutTask, err := q.CreateDbTask(h.DbApi.ConnCtx, dbTaskParams)

	tx.Commit(h.DbApi.ConnCtx)

	job_ := &job.BaseJob{ID: r.JobId}
	job_.Tasks = append(job_.Tasks, db_task.NewDbTask(&migrateOutTask, h.DbApi))

	h.JobProducer.Send(job_)

	return nil
}
