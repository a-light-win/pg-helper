package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type MigrateDbRequest struct {
	DbName        string `json:"db_name" binding:"required,max=63,id"`
	BackupPath    string `json:"backup_path" binding:"max=255"`
	BackupVersion int    `json:"backup_version"`

	Config *config.DbConfig `json:"-"`
	tx     pgx.Tx           `json:"-"`
	Query  *db.Queries      `json:"-"`
}

func (h *Handler) MigrateDb(c *gin.Context) {
	req, err := h.checkMigrateDbRequest(c)
	if err != nil {
		return
	}
	defer req.tx.Rollback(c)

	// Enusre the database exists
	db_, err := req.Query.GetDbByName(c, req.DbName)
	if err != nil {
		if err == pgx.ErrNoRows {
			logErrorAndRespondWithCode(c, err, "Database not found", http.StatusNotFound)
		} else {
			logErrorAndRespond(c, err, "Get database failed")
		}
		return
	}

	// If there is an active migration task for the database,
	// return the task immidiately
	task, err := req.Query.GetActiveDbTaskByDbID(c,
		db.GetActiveDbTaskByDbIDParams{DbID: db_.ID, Action: db.DbActionMigrate})
	if err != nil {
		if err != pgx.ErrNoRows {
			logErrorAndRespond(c, err, "Get active migration task failed")
			return
		}
	} else {
		c.JSON(http.StatusOK, task)
		return
	}

	// Ensuere database is in enabled state
	if !db_.IsEnabled {
		err := fmt.Errorf("database is not enabled")
		logErrorAndRespondWithCode(c, err, "Database is not enabled", http.StatusBadRequest)
		return
	}

	// and there is no tables created in the database
	// We do not restore to a database with data, as it may cause data loss
	if err = checkDbIsEmpty(c, req.Config, req.DbName); err != nil {
		return
	}

	// Create new migration tasks

	// 1. task to request remote pg-helper to backup the database
	var remoteBackupTask db.DbTask
	data := dto.DbTaskData{BackupPath: req.BackupPath}
	if req.BackupPath == "" {
		// TODO: init the backup path here
		remoteBackupTask, err = req.Query.CreateDbTask(c, db.CreateDbTaskParams{
			DbID: db_.ID, Action: db.DbActionRemoteBackup, Reason: "Migrate database",
			Status: db.DbTaskStatusPending, Data: data,
		})
		if err != nil {
			logErrorAndRespond(c, err, "Create remote backup task failed")
			return
		}
		data.DependsOn = append(data.DependsOn, remoteBackupTask.ID)
	}

	// 2. task to restore the backup to the new database
	restoreTask, err := req.Query.CreateDbTask(c, db.CreateDbTaskParams{
		DbID: db_.ID, Action: db.DbActionRestore, Reason: "Migrate database",
		Status: db.DbTaskStatusPending, Data: data,
	})
	if err != nil {
		logErrorAndRespond(c, err, "Create restore task failed")
		return
	}

	// 3. done the migrate task
	data.DependsOn = append(data.DependsOn, restoreTask.ID)
	migrateTask, err := req.Query.CreateDbTask(c, db.CreateDbTaskParams{
		DbID: db_.ID, Action: db.DbActionMigrate, Reason: "Migrate database",
		Status: db.DbTaskStatusPending, Data: data,
	})
	if err != nil {
		logErrorAndRespond(c, err, "Create migrate task failed")
		return
	}

	req.tx.Commit(c)

	if req.BackupPath == "" {
		h.JobProducer.Produce(&remoteBackupTask)
	}
	h.JobProducer.Produce(&restoreTask)
	readyChan := h.JobProducer.Produce(&migrateTask)

	select {
	case <-readyChan:
	case <-time.After(5 * time.Second):
	}

	q := req.Query.WithTx(req.tx)
	task, err = q.GetTaskByID(c, migrateTask.ID)
	if err != nil {
		log.Warn().Err(err).Msg("Get task by ID failed, return the task without status update")
		c.JSON(http.StatusOK, migrateTask)
		return
	}
	c.JSON(http.StatusOK, task)
}

func (h *Handler) checkMigrateDbRequest(c *gin.Context) (*MigrateDbRequest, error) {
	req := &MigrateDbRequest{}
	if err := c.ShouldBind(req); err != nil {
		logErrorAndRespondWithCode(c, err, "Check migrade db request failed", http.StatusBadRequest)
		return nil, err
	}

	if isReservedDbName(req.DbName) {
		err := fmt.Errorf("the database name is Reserved")
		logErrorAndRespondWithCode(c, err, "The database name is Reserved", http.StatusBadRequest)
		return nil, err
	}

	if req.BackupPath != "" {
		if pgVersion, err := req.Config.ValidateBackupPath(req.BackupPath, req.DbName); err != nil {
			logErrorAndRespondWithCode(c, err, "Invalid backup path", http.StatusBadRequest)
			return nil, err
		} else {
			req.BackupVersion = pgVersion
		}
	}

	// TODO: auth check

	conn := c.MustGet("DbConn").(*pgxpool.Conn)
	if tx, err := conn.Begin(c); err != nil {
		logErrorAndRespond(c, err, "Begin transaction failed")
		return nil, err
	} else {
		req.tx = tx
	}
	req.Query = db.New(req.tx)

	return req, nil
}

var ReservedDbNames = []string{"template0", "template1", "postgres"}

func isReservedDbName(dbName string) bool {
	for _, reservedDbName := range ReservedDbNames {
		if dbName == reservedDbName {
			return true
		}
	}
	return false
}

func checkDbIsEmpty(c *gin.Context, config *config.DbConfig, dbName string) error {
	conn, err := pgx.Connect(c, config.Url(dbName))
	if err != nil {
		logErrorAndRespond(c, err, "Connect to database failed")
		return err
	}
	defer conn.Close(c)

	q := db.New(conn)
	if count, err := q.CountDbTables(c); err != nil {
		logErrorAndRespond(c, err, "Count tables failed")
		return err
	} else {
		if count > 0 {
			err := fmt.Errorf("database is not empty")
			return err
		}
		return nil
	}
}
