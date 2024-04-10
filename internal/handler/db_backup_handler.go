package handler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/a-light-win/pg-helper/internal/config"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type BackupDbRequest struct {
	dto.BackupDbRequest

	query *db.Queries `json:"-"`
	db    *db.Db      `json:"-"`
}

type BackupDbResponse struct {
	ID       uuid.UUID        `json:"id"`
	DbID     int64            `json:"db_id"`
	DbName   string           `json:"db_name"`
	Action   db.DbAction      `json:"action"`
	Reason   string           `json:"reason"`
	Status   db.DbTaskStatus  `json:"status"`
	CreateAt pgtype.Timestamp `json:"created_at"`
	UpdateAt pgtype.Timestamp `json:"updated_at"`
	Data     dto.DbTaskData   `json:"data"`
}

func (h *Handler) BackupDb(c *gin.Context) {
	// Check if the request is valid
	request, err := checkBackupDbRequest(c)
	if err != nil {
		return
	}

	// Check if the database exists
	request.db, err = mustGetDbByName(c, request.query, request.Name)
	if err != nil {
		return
	}

	// Ensure there is only one backup tasks for the database at a time
	// This is to prevent the database from being backed up multiple times concurrently
	backupTask, err := getBackupTask(c, request.query, request.db.ID)
	if err != nil {
		return
	}
	if backupTask != nil {
		resp := fromDbTask(request.db.Name, backupTask)
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
		Data:     task.Data,
	}
	return &resp
}

func mustGetDbByName(c *gin.Context, query *db.Queries, dbName string) (*db.Db, error) {
	db, err := query.GetDbByName(c, dbName)
	if err != nil {
		if err == pgx.ErrNoRows {
			logErrorAndRespondWithCode(c, err, "Database not found", http.StatusNotFound)
		} else {
			logErrorAndRespond(c, err, "Failed to get database")
		}
		return nil, err
	}
	return &db, nil
}

func checkBackupDbRequest(c *gin.Context) (*BackupDbRequest, error) {
	request := BackupDbRequest{}
	if err := c.ShouldBind(&request); err != nil {
		logErrorAndRespondWithCode(c, err, "Failed to bind request", http.StatusBadRequest)
		return nil, err
	}
	request.query = db.New(c.MustGet("DbConn").(*pgxpool.Conn))
	return &request, nil
}

func getBackupTask(c *gin.Context, query *db.Queries, dbID int64) (*db.DbTask, error) {
	backupTask, err := query.GetActiveDbTaskByDbID(c, db.GetActiveDbTaskByDbIDParams{DbID: dbID, Action: db.DbActionBackup})
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
	_backupData := dto.DbTaskData{BackupPath: h.Config.Db.NewBackupFile(request.Name)}

	task, err := request.query.CreateDbTask(c, db.CreateDbTaskParams{
		DbID:   request.db.ID,
		Action: db.DbActionBackup,
		Reason: request.Reason,
		Status: db.DbTaskStatusRunning,
		Data:   _backupData,
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
	TaskID     uuid.UUID
}

func (h *DbBackupHandler) backup() {
	// Ensuere backup dir is exists
	os.MkdirAll(h.Config.BackupDbDir(h.DbName), 0750)

	// Backup the database here
	args := []string{
		"-h", h.Config.Host(h.Config.CurrentVersion),
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
