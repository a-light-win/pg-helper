package grpc_agent

import (
	"fmt"

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

type CreateDatabaseRequest struct {
	*proto.CreateDatabaseJob
	JobId uuid.UUID
}

func NewCreateDatabaseRequest(job *proto.DbJob) *CreateDatabaseRequest {
	return &CreateDatabaseRequest{
		CreateDatabaseJob: job.GetCreateDatabase(),
		JobId:             utils.StringToUuid(job.JobId),
	}
}

func (r *CreateDatabaseRequest) Process(h *GrpcAgentHandler) error {
	if err := r.validate(h); err != nil {
		// TODO: notify the grpc server
		return err
	}

	return h.DbApi.QueryWithRollback(func(tx pgx.Tx) error {
		return r.process(h, tx)
	})
}

func (r *CreateDatabaseRequest) validate(h *GrpcAgentHandler) error {
	if h.DbApi.DbConfig.IsReservedName(r.Name) {
		err := fmt.Errorf("the database name is Reserved")
		return err
	}
	return nil
}

func (r *CreateDatabaseRequest) process(h *GrpcAgentHandler, tx pgx.Tx) error {
	log := log.With().
		Str("DbName", r.Name).
		Str("Owner", r.Owner).
		Str("Reason", r.Reason).
		Str("MigrationFrom", r.MigrateFrom).
		Logger()

	dbApi := h.DbApi
	q := db.New(tx)

	database, err := dbApi.GetDbByName(r.Name, q)
	if err != nil && err != pgx.ErrNoRows {
		log.Warn().Err(err).Msg("Get database record faield")
		// TODO: notify grpc server
		return logger.NewAlreadyLoggedError(err, zerolog.WarnLevel)
	}

	if database == nil {
		params := &db.CreateDbParams{Name: r.Name, Owner: r.Owner}
		database, err = dbApi.CreateDb(params, q)
		if err != nil {
			return err
		}
	}

	if !database.IsNotExist() {
		// if JobId exists and equal to the LastJobId,
		// loading the previous failed/cancelled tasks and retry
		// else new a JobId, and create new tasks.
		if r.JobId != uuid.Nil && r.JobId == database.LastJobID {
			// TODO: Load last failed job and retry it from the first failed task
			return nil
		}
	}

	if r.JobId == uuid.Nil {
		r.JobId = uuid.New()
	}

	dbTaskParams := &db.CreateDbTaskParams{
		JobID:  r.JobId,
		DbID:   database.ID,
		DbName: database.Name,
		Reason: r.Reason,
		Action: db.DbActionCreateUser,
		Status: db.DbTaskStatusPending,
		Data: db.DbTaskData{
			Owner:      r.Owner,
			BackupFrom: r.MigrateFrom,
			BackupPath: h.DbApi.DbConfig.NewBackupFile(r.Name),
		},
	}

	job_ := &job.BaseJob{
		ID:   r.JobId,
		Name: fmt.Sprintf("CreateDatabase-%s", r.Name),
	}

	createUserTask, err := dbApi.CreateDbTask(dbTaskParams, q)
	if err != nil {
		return err
	}
	createUserTask.Data.Password = r.Password
	job_.Tasks = append(job_.Tasks, db_task.NewDbTask(createUserTask, dbApi))

	dbTaskParams.Action = db.DbActionCreate
	dbTaskParams.Data.DependsOn = []uuid.UUID{createUserTask.ID}
	createDbTask, err := dbApi.CreateDbTask(dbTaskParams, q)
	if err != nil {
		return err
	}
	job_.Tasks = append(job_.Tasks, db_task.NewDbTask(createDbTask, dbApi))

	if r.MigrateFrom != "" {
		dbTaskParams.Action = db.DbActionBackup
		dbTaskParams.Data.DependsOn = []uuid.UUID{createDbTask.ID}
		backupDbTask, err := dbApi.CreateDbTask(dbTaskParams, q)
		if err != nil {
			return err
		}
		job_.Tasks = append(job_.Tasks, db_task.NewDbTask(backupDbTask, dbApi))
	}

	if r.MigrateFrom != "" || r.BackupPath != "" {
		dbTaskParams.Action = db.DbActionRestore
		dbTaskParams.Data.DependsOn = []uuid.UUID{job_.Tasks[len(job_.Tasks)-1].UUID()}
		restoreDbTask, err := dbApi.CreateDbTask(dbTaskParams, q)
		if err != nil {
			return err
		}
		job_.Tasks = append(job_.Tasks, db_task.NewDbTask(restoreDbTask, dbApi))
	}

	dbTaskParams.Action = db.DbActionWaitReady
	dbTaskParams.Data.DependsOn = []uuid.UUID{job_.Tasks[len(job_.Tasks)-1].UUID()}
	waitReadyTask, err := dbApi.CreateDbTask(dbTaskParams, q)
	if err != nil {
		return err
	}
	job_.Tasks = append(job_.Tasks, db_task.NewDbTask(waitReadyTask, dbApi))

	tx.Commit(dbApi.ConnCtx)

	h.JobProducer.Send(job_)
	return nil
}
