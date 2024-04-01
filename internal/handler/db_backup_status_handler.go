package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (h *Handler) BackupStatus(c *gin.Context) {
	var taskId pgtype.UUID
	if err := taskId.Scan(c.Param("taskId")); err != nil {
		logErrorAndRespond(c, err, "invalid task ID")
		return
	}

	query := db.New(c.MustGet("DbConn").(*pgxpool.Conn))

	task, err := query.GetTaskByID(c, taskId)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		logErrorAndRespond(c, err, "failed to get task")
		return
	}

	dbName := ""
	if db_, err := query.GetDbByID(c, task.DbID); err == nil {
		dbName = db_.Name
	}

	resp := fromDbTask(dbName, &task)

	if resp.Status == db.DbTaskStatusRunning {
		filePath := filepath.Join(h.Config.Db.BackupRootPath, resp.Data.BackupPath)
		newStatus := db.DbTaskStatusRunning
		if _, err := os.Stat(filePath); err == nil {
			newStatus = db.DbTaskStatusCompleted
		} else if _, err := os.Stat(filePath + ".error"); err == nil {
			newStatus = db.DbTaskStatusFailed
		}

		if resp.Status != newStatus {
			resp.Status = newStatus
			query.SetDbTaskStatus(c, db.SetDbTaskStatusParams{Status: newStatus, ID: task.ID})
		}
	}

	c.JSON(http.StatusOK, resp)
}
