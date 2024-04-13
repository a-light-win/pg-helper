package grpc_handler

import (
	"errors"
	"fmt"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/job/db_job"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type CreateDatabaseRequest struct {
	TaskId string `validate:"required,uuid4"`
	Reason string `validate:"required,max=255"`

	Name        string `validate:"required,max=63,id"`
	Owner       string `validate:"required,max=63,id"`
	Password    string `validate:"min=8"`
	MigrateFrom int32  `validate:"pg_ver"`

	log zerolog.Logger
}

func NewCreateDatabaseHandler(task *proto.DbTask) *CreateDatabaseRequest {
	taskData := task.GetCreateDatabase()
	r := &CreateDatabaseRequest{
		TaskId: task.TaskId,
		Reason: task.Reason,

		Name:        taskData.Name,
		Owner:       taskData.Owner,
		Password:    taskData.Password,
		MigrateFrom: taskData.MigrateFrom,
	}

	r.log = log.With().
		Str("DbName", r.Name).
		Str("Owner", r.Owner).
		Int32("MigrateFrom", r.MigrateFrom).
		Logger()

	return r
}

func (r *CreateDatabaseRequest) Validate() error {
	if err := Validator.Struct(r); err != nil {
		r.log.Warn().Err(err).Msg("Validation failed")
		return err
	}

	if gd_.DbConfig.IsReservedName(r.Name) {
		err := fmt.Errorf("the database name is Reserved")
		r.log.Warn().Err(err).Msg("The database name is Reserved")
		return err
	}
	return nil
}

func (r *CreateDatabaseRequest) Handle() error {
	conn, err := gd_.DbPool.Acquire(gd_.ConnCtx)
	if err != nil {
		return err
	}
	defer conn.Release()

	if err := r.createUser(conn); err != nil {
		return err
	}

	database, err := r.createDb(conn)
	if err != nil {
		return err
	}

	return r.createBackgroundJob(conn, database)
}

func (r *CreateDatabaseRequest) OnError(err error) {
	// TODO: Notify grpc server the task is failed
}

func (r *CreateDatabaseRequest) createUser(conn *pgxpool.Conn) error {
	q := db.New(conn)
	_, err := q.IsUserExists(gd_.ConnCtx, pgtype.Text{String: r.Owner, Valid: true})
	if err != nil {
		if err != pgx.ErrNoRows {
			r.log.Warn().Err(err).Msg("Failed to check if user exists")
			return err
		}

		// Create User here
		pgconn := conn.Conn().PgConn()
		escapedPassword, err := pgconn.EscapeString(r.Password)
		if err != nil {
			r.log.Warn().Err(err).Msg("Failed to escape password")
			return err
		}

		_, err = conn.Exec(gd_.ConnCtx, fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'",
			r.Owner,
			escapedPassword))
		if err != nil {
			r.log.Warn().Err(err).Msg("Failed to create user")
			return err
		}

		r.log.Log().Msg("User created successfully")
	}

	return nil
}

func (r *CreateDatabaseRequest) createDb(conn *pgxpool.Conn) (*db.Db, error) {
	q := db.New(conn)
	database, err := q.GetDbByName(gd_.ConnCtx, r.Name)
	if err != nil {
		if err != pgx.ErrNoRows {
			r.log.Warn().Err(err).Msg("Failed to check if database exists")
			return nil, err
		}

		database, err = q.CreateDb(gd_.ConnCtx, db.CreateDbParams{Name: r.Name, Owner: r.Owner})
		if err != nil {
			r.log.Warn().Err(err).Msg("Failed to create database")
			return nil, err
		}
	} else {
		if database.Owner != r.Owner {
			err := errors.New("database exists with another owner")
			r.log.Error().Err(err).
				Str("actualOwner", database.Owner).
				Msg("")
			return nil, err
		}
	}

	if database.Status == proto.DbStatus_NotExist {
		// Create database
		_, err := conn.Exec(gd_.ConnCtx, fmt.Sprintf("CREATE DATABASE %s OWNER %s",
			r.Name, r.Owner))
		if err != nil {
			r.log.Warn().Err(err).Msg("Failed to create database")
			return nil, err
		}

		database.Status = proto.DbStatus_Created
		q.SetDbStatus(gd_.ConnCtx, db.SetDbStatusParams{ID: database.ID, Status: database.Status})

		r.log.Log().Msg("Database created successfully")
	}

	if database.Status == proto.DbStatus_Created {
		if r.MigrateFrom == 0 {
			database.Status = proto.DbStatus_Ready
			q.SetDbStatus(gd_.ConnCtx, db.SetDbStatusParams{ID: database.ID, Status: database.Status})
			r.log.Debug().Msg("Database is ready because no migration is needed")
		}
	}

	return &database, nil
}

func (r *CreateDatabaseRequest) createBackgroundJob(conn *pgxpool.Conn, database *db.Db) error {
	switch database.Status {
	case proto.DbStatus_Backuping, proto.DbStatus_Restoring, proto.DbStatus_Ready:
		return nil
	case proto.DbStatus_Migrated:
		err := errors.New("database is already migrate to another instance")
		return err
	case proto.DbStatus_Expired:
		err := errors.New("database is expired")
		return err
	}

	q := db.New(conn)
	_, err := q.GetActiveDbTaskByDbID(gd_.ConnCtx,
		db.GetActiveDbTaskByDbIDParams{DbID: database.ID, Action: db.DbActionWaitReady})
	if err != nil {
		if err != pgx.ErrNoRows {
			r.log.Warn().Err(err).Msg("Get active migrate task failed")
			return err
		}
	} else {
		// there is already an active task to migrate the database
		return nil
	}

	// and there is no tables created in the database
	// We do not restore to a database with data, as it may cause data loss
	if err = r.checkDbIsEmpty(); err != nil {
		return err
	}

	tx, err := conn.Begin(gd_.ConnCtx)
	if err != nil {
		r.log.Warn().Err(err).Msg("Failed to begin transaction")
		return err
	}
	defer tx.Rollback(gd_.ConnCtx)

	q = db.New(tx)

	dbTaskParams := db.CreateDbTaskParams{
		DbID:   database.ID,
		Action: db.DbActionBackup,
		Reason: r.Reason,
		Status: db.DbTaskStatusPending,
		Data:   db.DbTaskData{BackupPath: gd_.DbConfig.NewBackupFile(r.Name)},
	}
	data := &dbTaskParams.Data

	backupTask, err := q.CreateDbTask(gd_.ConnCtx, dbTaskParams)
	if err != nil {
		r.log.Warn().Err(err).Msg("Create backup task failed")
		return err
	}

	dbTaskParams.Action = db.DbActionRestore
	data.DependsOn = []uuid.UUID{backupTask.ID}
	restoreTask, err := q.CreateDbTask(gd_.ConnCtx, dbTaskParams)
	if err != nil {
		r.log.Warn().Err(err).Msg("Create restore task failed")
		return err
	}

	dbTaskParams.Action = db.DbActionWaitReady
	data.DependsOn = append(data.DependsOn, restoreTask.ID)
	waitReadyTask, err := q.CreateDbTask(gd_.ConnCtx, dbTaskParams)
	if err != nil {
		r.log.Warn().Err(err).Msg("Create wait ready task failed")
		return err
	}

	tx.Commit(gd_.ConnCtx)

	gd_.JobProducer.Produce(db_job.NewDbJob(&backupTask))
	gd_.JobProducer.Produce(db_job.NewDbJob(&restoreTask))
	gd_.JobProducer.Produce(db_job.NewDbJob(&waitReadyTask))
	return nil
}

func (r *CreateDatabaseRequest) checkDbIsEmpty() error {
	conn, err := pgx.Connect(gd_.ConnCtx, gd_.DbConfig.Url(r.Name, 0))
	if err != nil {
		r.log.Warn().Err(err).Msg("Failed to connect to the database")
		return err
	}
	defer conn.Close(gd_.ConnCtx)

	q := db.New(conn)
	if count, err := q.CountDbTables(gd_.ConnCtx); err != nil {
		r.log.Warn().Err(err).Msg("Failed to count tables in the database")
		return err
	} else {
		if count > 0 {
			err := fmt.Errorf("database is not empty")
			r.log.Warn().Err(err).Msg("Database is not empty")
			return err
		}
		return nil
	}
}
