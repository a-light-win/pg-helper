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
		log.Warn().Err(err).Msg("Validation failed")
		return err
	}

	if h.DbApi.DbConfig.IsReservedName(request.Name) {
		err := fmt.Errorf("the database name is Reserved")
		log.Warn().Err(err).Msg("The database name is Reserved")
		return err
	}
	return nil
}

func (h *GrpcAgentHandler) createDatabase(task *proto.DbTask) error {
	request := newCreateDatabaseRequest(task)
	if err := h.validateCreateDatabaseRequest(request); err != nil {
		return err
	}

	conn, err := h.DbApi.Acquire()
	if err != nil {
		return err
	}

	defer conn.Release()

	if err := h.createUser(conn, &request.CreateOwnerRequest); err != nil {
		return err
	}

	database, err := h.createDb(conn, request)
	if err != nil {
		return err
	}

	return h.createBackgroundJob(conn, request, database)
}

func (r *CreateDatabaseRequest) OnError(err error) {
	// TODO: Notify grpc server the task is failed
}

func (h *GrpcAgentHandler) createUser(conn *pgxpool.Conn, request *CreateOwnerRequest) error {
	q := db.New(conn)
	connCtx := h.DbApi.ConnCtx

	_, err := q.IsUserExists(connCtx, pgtype.Text{String: request.Owner, Valid: true})
	if err != nil {
		if err != pgx.ErrNoRows {
			log.Warn().Err(err).Msg("Failed to check if user exists")
			return err
		}

		// Create User here
		pgconn := conn.Conn().PgConn()
		escapedPassword, err := pgconn.EscapeString(request.Password)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to escape password")
			return err
		}

		_, err = conn.Exec(connCtx, fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'",
			request.Owner,
			escapedPassword))
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create user")
			return err
		}

		log.Log().Msg("User created successfully")
	}

	return nil
}

func (h *GrpcAgentHandler) createDb(conn *pgxpool.Conn, request *CreateDatabaseRequest) (*db.Db, error) {
	q := db.New(conn)
	connCtx := h.DbApi.ConnCtx
	database, err := q.GetDbByName(connCtx, request.Name)
	if err != nil {
		if err != pgx.ErrNoRows {
			log.Warn().Err(err).Msg("Failed to check if database exists")
			return nil, err
		}

		database, err = q.CreateDb(connCtx, db.CreateDbParams{Name: request.Name, Owner: request.Owner})
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create database")
			return nil, err
		}
	} else {
		if database.Owner != request.Owner {
			err := errors.New("database exists with another owner")
			log.Error().Err(err).
				Str("actualOwner", database.Owner).
				Msg("")
			return nil, err
		}
	}

	if database.Stage == proto.DbStage_None {

		database.Stage = proto.DbStage_Creating
		database.Status = proto.DbStatus_Processing

		dbStatusParams := db.SetDbStatusParams{ID: database.ID, Status: database.Status, Stage: database.Stage}
		if err := q.SetDbStatus(connCtx, dbStatusParams); err != nil {
			log.Warn().Err(err).Msg("Failed to set database status")
			return nil, err
		}

		// Create database
		_, err := conn.Exec(connCtx, fmt.Sprintf("CREATE DATABASE %s OWNER %s",
			request.Name, request.Owner))
		if err != nil {
			log.Warn().Err(err).Msg("Failed to create database")
			return nil, err
		}

		database.Status = proto.DbStatus_Done

		dbStatusParams.Status = database.Status
		if err := q.SetDbStatus(connCtx, dbStatusParams); err != nil {
			log.Warn().Err(err).Msg("Failed to set database status")
			return nil, err
		}

		log.Log().Msg("Database created successfully")
	}

	if database.Stage == proto.DbStage_Creating && database.Status == proto.DbStatus_Done {
		if request.MigrateFrom == "" {
			database.Stage = proto.DbStage_Running
			q.SetDbStatus(connCtx, db.SetDbStatusParams{ID: database.ID, Status: database.Status, Stage: database.Stage})
			log.Debug().Msg("Database is ready because no migration is needed")
		}
	}

	return &database, nil
}

func (h *GrpcAgentHandler) createBackgroundJob(conn *pgxpool.Conn, request *CreateDatabaseRequest, database *db.Db) error {
	switch database.Stage {
	case proto.DbStage_Backuping, proto.DbStage_Restoring, proto.DbStage_Running:
		return nil
	case proto.DbStage_MigrateOut:
		err := errors.New("database is already migrate to another instance")
		return err
	case proto.DbStage_Dropping:
		err := errors.New("database is dropping")
		return err
	}

	q := db.New(conn)
	connCtx := h.DbApi.ConnCtx

	_, err := q.GetActiveDbTaskByDbID(connCtx,
		db.GetActiveDbTaskByDbIDParams{DbID: database.ID, Action: db.DbActionWaitReady})
	if err != nil {
		if err != pgx.ErrNoRows {
			log.Warn().Err(err).Msg("Get active migrate task failed")
			return err
		}
	} else {
		// there is already an active task to migrate the database
		return nil
	}

	// and there is no tables created in the database
	// We do not restore to a database with data, as it may cause data loss
	if err = h.checkDbIsEmpty(request.Name); err != nil {
		return err
	}

	tx, err := conn.Begin(connCtx)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to begin transaction")
		return err
	}
	defer tx.Rollback(connCtx)

	q = db.New(tx)

	dbTaskParams := db.CreateDbTaskParams{
		DbID:   database.ID,
		Action: db.DbActionBackup,
		Reason: request.Reason,
		Status: db.DbTaskStatusPending,
		Data:   db.DbTaskData{BackupFrom: request.MigrateFrom, BackupPath: h.DbApi.DbConfig.NewBackupFile(request.Name)},
	}
	data := &dbTaskParams.Data

	backupTask, err := q.CreateDbTask(connCtx, dbTaskParams)
	if err != nil {
		log.Warn().Err(err).Msg("Create backup task failed")
		return err
	}

	dbTaskParams.Action = db.DbActionRestore
	data.DependsOn = []uuid.UUID{backupTask.ID}
	restoreTask, err := q.CreateDbTask(connCtx, dbTaskParams)
	if err != nil {
		log.Warn().Err(err).Msg("Create restore task failed")
		return err
	}

	dbTaskParams.Action = db.DbActionWaitReady
	data.DependsOn = append(data.DependsOn, restoreTask.ID)
	waitReadyTask, err := q.CreateDbTask(connCtx, dbTaskParams)
	if err != nil {
		log.Warn().Err(err).Msg("Create wait ready task failed")
		return err
	}

	tx.Commit(connCtx)

	h.JobProducer.Produce(db_job.NewDbJob(&backupTask))
	h.JobProducer.Produce(db_job.NewDbJob(&restoreTask))
	h.JobProducer.Produce(db_job.NewDbJob(&waitReadyTask))
	return nil
}

func (h *GrpcAgentHandler) checkDbIsEmpty(name string) error {
	connCtx := h.DbApi.ConnCtx
	dbConfig := h.DbApi.DbConfig
	conn, err := pgx.Connect(connCtx, dbConfig.Url(name, nil))
	if err != nil {
		log.Warn().Err(err).Msg("Failed to connect to the database")
		return err
	}
	defer conn.Close(connCtx)

	q := db.New(conn)
	if count, err := q.CountDbTables(connCtx); err != nil {
		log.Warn().Err(err).Msg("Failed to count tables in the database")
		return err
	} else {
		if count > 0 {
			err := fmt.Errorf("database is not empty")
			log.Warn().Err(err).Msg("Database is not empty")
			return err
		}
		return nil
	}
}
