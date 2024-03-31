package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

type CreateDbRequest struct {
	Name     string `json:"name" binding:"required,max=63,id"`
	Owner    string `json:"owner" binding:"required,max=63,id"`
	Password string `json:"password" binding:"required,min=8"`

	conn  *pgxpool.Conn `json:"-"`
	query *db.Queries   `json:"-"`
}

func (h *Handler) CreateDb(c *gin.Context) {
	request, err := checkCreateDbRequest(c)
	if err != nil {
		return
	}

	if exists, err := isDbExists(c, request); exists || err != nil {
		return
	}

	// Enusre the user exists
	if err := createUser(c, request); err != nil {
		return
	}

	// Create Database here
	createDb(c, request)
}

func checkCreateDbRequest(c *gin.Context) (*CreateDbRequest, error) {
	request := CreateDbRequest{}
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to bind request", "detail": err.Error()})
		return nil, err
	}

	request.conn = c.MustGet("DbConn").(*pgxpool.Conn)
	request.query = db.New(request.conn)
	return &request, nil
}

func createUser(c *gin.Context, request *CreateDbRequest) error {
	_, err := request.query.IsUserExists(c, pgtype.Text{String: request.Owner, Valid: true})
	if err != nil {
		if err != pgx.ErrNoRows {
			logErrorAndRespond(c, err, "Failed to check if user exists")
			return err
		}

		// Create User here
		pgconn := request.conn.Conn().PgConn()
		escapedPassword, err := pgconn.EscapeString(request.Password)
		if err != nil {
			logErrorAndRespondWithCode(c, err, "Failed to escape password", http.StatusBadRequest)
			return err
		}
		_, err = request.conn.Exec(c, fmt.Sprintf("CREATE USER %s WITH PASSWORD '%s'",
			request.Owner,
			escapedPassword))
		if err != nil {
			logErrorAndRespond(c, err, "Failed to create user")
			return err
		}

		log.Log().Str("user", request.Owner).Msg("User created successfully")
	}

	return nil
}

func isDbExists(c *gin.Context, request *CreateDbRequest) (bool, error) {
	owner, err := request.query.GetDbOwner(c, request.Name)
	if err != nil {
		if err != pgx.ErrNoRows {
			logErrorAndRespond(c, err, "Failed to get database owner")
			return false, err
		}
	}

	if owner.Valid {
		if owner.String == request.Owner {
			c.JSON(http.StatusOK, gin.H{"message": "Database already exists"})
			return true, nil
		}

		log.Error().
			Str("db_name", request.Name).
			Str("owner", owner.String).
			Str("request_owner", request.Owner).
			Msg("Database exists with another owner")

		c.JSON(http.StatusBadRequest, gin.H{"error": "Database exists with another owner"})
		return false, errors.New("database exists with another owner")
	}
	return false, nil
}

func createDb(c *gin.Context, request *CreateDbRequest) error {
	_, err := request.conn.Exec(c, fmt.Sprintf("CREATE DATABASE %s OWNER %s",
		request.Name, request.Owner))
	if err != nil {
		logErrorAndRespond(c, err, "Failed to create database")
		return err
	}

	log.Log().Msg("Database created successfully")
	c.JSON(http.StatusCreated, gin.H{"message": "Database created successfully"})
	return nil
}
