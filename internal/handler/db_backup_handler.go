package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type BackupDbRequest struct {
	Name   string `json:"name" binding:"required,max=63,id"`
	Reason string `json:"reason" binding:"required,max=255"`

	query *db.Queries `json:"-"`
	db    *db.Db      `json:"-"`
}

type BackupDbResponse struct {
	ID       pgtype.UUID      `json:"id"`
	DbID     int64            `json:"db_id"`
	DbName   string           `json:"db_name"`
	Action   db.DbAction      `json:"action"`
	Reason   string           `json:"reason"`
	Status   db.DbTaskStatus  `json:"status"`
	CreateAt pgtype.Timestamp `json:"created_at"`
	UpdateAt pgtype.Timestamp `json:"updated_at"`
	Data     BAckupTaskData   `json:"data"`
}

type BAckupTaskData struct {
	BackupPath string `json:"backup_path"`
}

func (h *Handler) BackupDb(c *gin.Context) {
	// Check if the request is valid
	request, err := checkBackupDbRequest(c)
	if err != nil {
		return
	}

	// Check if the database exists
	db := mustGetDbByName(c, request.query, request.Name)
	if db == nil {
		return
	}
	request.db = db

	// Ensure there is only one backup tasks for the database at a time
	// This is to prevent the database from being backed up multiple times concurrently
	backupTask, err := getBackupTask(c, request.query, db.ID)
	if err != nil {
		return
	}
	if backupTask != nil {
		resp := fromDbTask(db.Name, backupTask)
		c.JSON(http.StatusOK, resp)
		return
	}

	// Backup Database here
	h.backupDb(c, request)
}

func fromDbTask(dbName string, task *db.DbTask) *BackupDbResponse {
	resp := BackupDbResponse{
		ID:       task.ID,
		DbID:     task.DbID,
		DbName:   dbName,
		Action:   task.Action,
		Reason:   task.Reason,
		Status:   task.Status,
		CreateAt: task.CreatedAt,
		UpdateAt: task.UpdatedAt,
	}
	if task.Data != nil {
		json.Unmarshal(task.Data, &resp.Data)
	}
	return &resp
}

func mustGetDbByName(c *gin.Context, query *db.Queries, dbName string) *db.Db {
	db, err := query.GetDbByName(c, dbName)
	if err != nil {
		if err == pgx.ErrNoRows {
			logErrorAndRespondWithCode(c, err, "Database not found", http.StatusNotFound)
		} else {
			logErrorAndRespond(c, err, "Failed to get database")
		}
		return nil
	}
	return &db
}

func checkBackupDbRequest(c *gin.Context) (*BackupDbRequest, error) {
	request := BackupDbRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		logErrorAndRespondWithCode(c, err, "Failed to bind request", http.StatusBadRequest)
		return nil, err
	}
	request.query = db.New(c.MustGet("DbConn").(*pgxpool.Conn))
	return &request, nil
}

func getBackupTask(c *gin.Context, query *db.Queries, dbID int64) (*db.DbTask, error) {
	backupTask, err := query.GetActiveTaskByDbID(c, db.GetActiveTaskByDbIDParams{DbID: dbID, Action: db.DbActionBackup})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		logErrorAndRespond(c, err, "Failed to get backup task")
		return nil, err
	}
	return &backupTask, nil
}

func (h *Handler) backupDb(c *gin.Context, request *BackupDbRequest) {
	// Create a backup task
	_backupData := BAckupTaskData{BackupPath: h.Config.Db.NewBackupFile(request.Name)}
	data, err := json.Marshal(_backupData)
	if err != nil {
		logErrorAndRespond(c, err, "Failed to marshal backup data")
		return
	}

	task, err := request.query.CreateDbTask(c, db.CreateDbTaskParams{
		DbID:   request.db.ID,
		Action: db.DbActionBackup,
		Reason: request.Reason,
		Status: db.DbTaskStatusRunning,
		Data:   data,
	})
	if err != nil {
		logErrorAndRespond(c, err, "Failed to create backup task")
		return
	}

	resp := fromDbTask(request.Name, &task)

	backupHandler := DbBackupHandler{
		Config:     h.Config.Db,
		Pool:       h.DbPool,
		DbName:     request.Name,
		BackupPath: resp.Data.BackupPath,
		TaskID:     task.ID,
	}

	go backupHandler.backup()

	c.JSON(http.StatusOK, resp)
}

type DbBackupHandler struct {
	Config     config.DbConfig
	Pool       *pgxpool.Pool
	DbName     string
	BackupPath string
	TaskID     pgtype.UUID
}

func (h *DbBackupHandler) backup() {
	// Ensuere backup dir is exists
	os.MkdirAll(h.Config.BackupDbDir(h.DbName), 0750)

	// Backup the database here
	args := []string{
		"-h", h.Config.Host,
		"-p", fmt.Sprint(h.Config.Port),
		"-U", h.Config.User,
		"-d", h.DbName,
		"-f", h.BackupPath + ".tmp",
	}

	cmd := exec.Command("pg_dump", args...)
	cmd.Dir = h.Config.BackupRootPath
	cmd.Stdin = strings.NewReader(h.Config.Password() + "\n")

	if err := cmd.Run(); err != nil {
		log.Error().Err(err).
			Str("DbName", h.DbName).
			Str("BackupPath", h.BackupPath).
			Msg("Failed to backup database")

		os.Remove(filepath.Join(h.Config.BackupRootPath, h.BackupPath+".tmp"))

		h.updateTaskStatus(db.DbTaskStatusFailed)
		return
	}

	os.Rename(filepath.Join(h.Config.BackupRootPath, h.BackupPath+".tmp"),
		filepath.Join(h.Config.BackupRootPath, h.BackupPath))

	log.Log().Str("DbName", h.DbName).
		Str("BackupPath", h.BackupPath).
		Msg("Database backup completed")

	h.updateTaskStatus(db.DbTaskStatusCompleted)
}

func (h *DbBackupHandler) updateTaskStatus(status db.DbTaskStatus) {
	// Update the task status
	ctx := context.Background()
	conn, err := h.Pool.Acquire(ctx)
	taskID, _ := h.TaskID.Value()
	if err != nil {
		log.Warn().Err(err).
			Str("DbName", h.DbName).
			Str("BackupPath", h.BackupPath).
			Str("TaskID", taskID.(string)).
			Msg("Failed to acquire connection")

		if status == db.DbTaskStatusFailed {
			h.writeStatusToFile()
		}
		return
	}
	defer conn.Release()

	q := db.New(conn)
	if err := q.SetDbTaskStatus(ctx, db.SetDbTaskStatusParams{ID: h.TaskID, Status: status}); err != nil {
		log.Warn().Err(err).
			Str("DbName", h.DbName).
			Str("BackupPath", h.BackupPath).
			Str("TaskID", taskID.(string)).
			Msg("Failed to update backup task status")

		if status == db.DbTaskStatusFailed {
			h.writeStatusToFile()
		}
	}
}

func (h *DbBackupHandler) writeStatusToFile() {
	file, err := os.Create(filepath.Join(h.Config.BackupRootPath, h.BackupPath+".error"))
	if err != nil {
		log.Error().Err(err).
			Str("DbName", h.DbName).
			Str("BackupPath", h.BackupPath).
			Msg("Failed to write error status to file")
		return
	}
	file.Close()
}
