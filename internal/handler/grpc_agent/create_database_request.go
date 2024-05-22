package grpc_agent

import (
	"errors"
	"fmt"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/handler/db_job"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type CreateOwnerRequest struct {
	Owner    string `validate:"required,max=63,id"`
	Password string `validate:"min=8"`
}

type CreateDatabaseRequest struct {
	CreateOwnerRequest

	RequestId string `validate:"required,uuid4"`
	Reason    string `validate:"required,max=255"`

	Name        string `validate:"required,max=63,id"`
	MigrateFrom string `validate:"max=63,iname"`
}

func newCreateDatabaseRequest(task *proto.DbTask) *CreateDatabaseRequest {
	taskData := task.GetCreateDatabase()
	r := &CreateDatabaseRequest{
		RequestId: task.RequestId,
		Reason:    taskData.Reason,

		Name:        taskData.Name,
		MigrateFrom: taskData.MigrateFrom,

		CreateOwnerRequest: CreateOwnerRequest{
			Owner:    taskData.Owner,
			Password: taskData.Password,
		},
	}
	return r
}

func (h *GrpcAgentHandler) validateCreateDatabaseRequest(request *CreateDatabaseRequest) error {
	if err := h.Validator.Struct(request); err != nil {
		return err
	}

	if h.DbApi.DbConfig.IsReservedName(request.Name) {
		err := fmt.Errorf("the database name is Reserved")
		return err
	}
	return nil
}

func (h *GrpcAgentHandler) createDatabase(task *proto.DbTask) error {
	taskData := task.GetCreateDatabase()
	logger := log.With().
		Str("DbName", taskData.Name).
		Str("Owner", taskData.Owner).
		Str("Reason", taskData.Reason).
		Str("MigrationFrom", taskData.MigrateFrom).
		Logger()

	request := newCreateDatabaseRequest(task)
	if err := h.validateCreateDatabaseRequest(request); err != nil {
		logger.Warn().Err(err).Msg("Validation failed")
		return err
	}

	conn, err := h.DbApi.Acquire()
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to acquire connection")
		return err
	}

	defer conn.Release()

	if err := h.createUser(logger, conn, &request.CreateOwnerRequest); err != nil {
		return err
	}

	database, err := h.createDb(logger, conn, request)
	if err != nil {
		return err
	}

	return h.createMigrateJob(logger, conn, request, database)
}

func (r *CreateDatabaseRequest) OnError(err error) {
	// TODO: Notify grpc server the task is failed
}

func (h *GrpcAgentHandler) createUser(logger zerolog.Logger, conn *pgxpool.Conn, request *CreateOwnerRequest) error {
	q := db.New(conn)
	connCtx := h.DbApi.ConnCtx

	_, err := q.IsUserExists(connCtx, pgtype.Text{String: request.Owner, Valid: true})
	if err != nil {
		if err != pgx.ErrNoRows {
			logger.Warn().Err(err).Msg("Failed to check if user exists")
			return err
		}

		// Create User here
		pgconn := conn.Conn().PgConn()
		escapedPassword, err := pgconn.EscapeString(request.Password)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to escape password")
			return err
		}

		_, err = conn.Exec(connCtx, fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'",
			request.Owner,
			escapedPassword))
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to create user")
			return err
		}

		logger.Log().Msg("User created successfully")
	}

	return nil
}

func (h *GrpcAgentHandler) createDb(logger zerolog.Logger, conn *pgxpool.Conn, request *CreateDatabaseRequest) (*db.Db, error) {
	q := db.New(conn)
	connCtx := h.DbApi.ConnCtx

	database, err := h.DbApi.GetDbByName(request.Name, q)
	if err != nil {
		if err != pgx.ErrNoRows {
			logger.Warn().Err(err).Msg("Failed to check if database exists")
			return nil, err
		}

		database, err = h.DbApi.CreateDb(&db.CreateDbParams{Name: request.Name, Owner: request.Owner}, q)
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to create database")
			return nil, err
		}
	} else {
		if database.Owner != request.Owner {
			err := errors.New("database exists with another owner")
			logger.Error().Err(err).
				Str("actualOwner", database.Owner).
				Msg("")
			return nil, err
		}
	}

	if database.Stage == proto.DbStage_None {

		database.Stage = proto.DbStage_Creating
		database.Status = proto.DbStatus_Processing
		if err := h.DbApi.UpdateDbStatus(database, q); err != nil {
			logger.Warn().Err(err).Msg("Failed to set database status")
			return nil, err
		}

		// Create database
		_, err := conn.Exec(connCtx, fmt.Sprintf("CREATE DATABASE %s OWNER %s",
			request.Name, request.Owner))
		if err != nil {
			logger.Warn().Err(err).Msg("Failed to create database")
			return nil, err
		}

		database.Status = proto.DbStatus_Done
		if err := h.DbApi.UpdateDbStatus(database, q); err != nil {
			logger.Warn().Err(err).Msg("Failed to set database status")
			return nil, err
		}

		logger.Log().Msg("Database created successfully")
	}

	if database.Stage == proto.DbStage_Creating && database.Status == proto.DbStatus_Done {
		if request.MigrateFrom == "" {
			logger.Debug().Msg("Database is ready because no migration is needed")

			database.Stage = proto.DbStage_Running
			if err := h.DbApi.UpdateDbStatus(database, q); err != nil {
				logger.Warn().Err(err).Msg("Failed to set database status")
				return nil, err
			}
		}
	}

	return database, nil
}

func (h *GrpcAgentHandler) createMigrateJob(logger zerolog.Logger, conn *pgxpool.Conn, request *CreateDatabaseRequest, database *db.Db) error {
	switch database.Stage {
	case proto.DbStage_None:
		err := errors.New("database is not created")
		logger.Warn().Err(err).Msg("")
		return err
	case proto.DbStage_Creating:
		if database.Status != proto.DbStatus_Done {
			err := errors.New("database is not ready")
			logger.Warn().Err(err).Msg("")
			return err
		}
	case proto.DbStage_Running:
		logger.Debug().Msg("Database is already running")
		return nil
	case proto.DbStage_MigrateOut:
		err := errors.New("database is already migrate to another instance")
		logger.Warn().Err(err).Msg("")
		return err
	case proto.DbStage_Dropping:
		err := errors.New("database is dropping")
		logger.Warn().Err(err).Msg("")
		return err
	}

	q := db.New(conn)
	connCtx := h.DbApi.ConnCtx

	_, err := q.GetActiveDbTaskByDbID(connCtx,
		db.GetActiveDbTaskByDbIDParams{DbID: database.ID, Action: db.DbActionWaitReady})
	if err != nil {
		if err != pgx.ErrNoRows {
			logger.Warn().Err(err).Msg("Get active migrate task failed")
			return err
		}
	} else {
		// there is already an active task to migrate the database
		logger.Debug().Msg("There is already an active task to migrate the database")
		return nil
	}

	// and there is no tables created in the database
	// We do not restore to a database with data, as it may cause data loss
	if err = h.checkDbIsEmpty(logger, request.Name); err != nil {
		return err
	}

	tx, err := conn.Begin(connCtx)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to begin transaction")
		return err
	}
	defer tx.Rollback(connCtx)

	q = db.New(tx)

	dbTaskParams := db.CreateDbTaskParams{
		DbID:   database.ID,
		DbName: database.Name,
		Action: db.DbActionBackup,
		Reason: request.Reason,
		Status: db.DbTaskStatusPending,
		Data:   db.DbTaskData{BackupFrom: request.MigrateFrom, BackupPath: h.DbApi.DbConfig.NewBackupFile(request.Name)},
	}
	data := &dbTaskParams.Data

	backupTask, err := q.CreateDbTask(connCtx, dbTaskParams)
	if err != nil {
		logger.Warn().Err(err).Msg("Create backup task failed")
		return err
	}

	dbTaskParams.Action = db.DbActionRestore
	data.DependsOn = []uuid.UUID{backupTask.ID}
	restoreTask, err := q.CreateDbTask(connCtx, dbTaskParams)
	if err != nil {
		logger.Warn().Err(err).Msg("Create restore task failed")
		return err
	}

	dbTaskParams.Action = db.DbActionWaitReady
	data.DependsOn = append(data.DependsOn, restoreTask.ID)
	waitReadyTask, err := q.CreateDbTask(connCtx, dbTaskParams)
	if err != nil {
		logger.Warn().Err(err).Msg("Create wait ready task failed")
		return err
	}

	tx.Commit(connCtx)

	h.JobProducer.Send(db_job.NewDbJob(&backupTask))
	h.JobProducer.Send(db_job.NewDbJob(&restoreTask))
	h.JobProducer.Send(db_job.NewDbJob(&waitReadyTask))
	return nil
}

func (h *GrpcAgentHandler) checkDbIsEmpty(logger zerolog.Logger, name string) error {
	connCtx := h.DbApi.ConnCtx
	dbConfig := h.DbApi.DbConfig
	conn, err := pgx.Connect(connCtx, dbConfig.Url(name, nil))
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to connect to the database")
		return err
	}
	defer conn.Close(connCtx)

	q := db.New(conn)
	if count, err := q.CountDbTables(connCtx); err != nil {
		logger.Warn().Err(err).Msg("Failed to count tables in the database")
		return err
	} else {
		if count > 0 {
			err := fmt.Errorf("database is not empty")
			logger.Warn().Err(err).Msg("Database is not empty")
			return err
		}
		return nil
	}
}
