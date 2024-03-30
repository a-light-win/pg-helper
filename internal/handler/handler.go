package handler

import (
	"fmt"
	"net/http"

	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

type Handler struct {
	DbPool *pgxpool.Pool
}

func New(dbPool *pgxpool.Pool) *Handler {
	return &Handler{DbPool: dbPool}
}

type CreateDbRequest struct {
	Name     string        `json:"name" binding:"required,max=63,id"`
	Owner    string        `json:"owner" binding:"required,max=63,id"`
	Password string        `json:"password" binding:"required"`
	Conn     *pgxpool.Conn `json:"-"`
}

func (h *Handler) CreateDb(c *gin.Context) {
	request := CreateDbRequest{}
	err := c.ShouldBindJSON(&request)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to bind request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to bind request"})
		return
	}

	request.Conn = c.MustGet("DbConn").(*pgxpool.Conn)

	q := db.New(request.Conn)
	owner, err := q.GetDbOwner(c, request.Name)
	if err != nil {
		if err != pgx.ErrNoRows {
			log.Error().
				Err(err).
				Msg("Failed to get database owner")

			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "Failed to get database owner"})

			return
		}
	}

	if owner.Valid {
		if owner.String == request.Owner {
			c.JSON(http.StatusOK, gin.H{"message": "Database already exists"})
			return
		}

		log.Error().
			Str("db_name", request.Name).
			Str("owner", owner.String).
			Str("request_owner", request.Owner).
			Msg("Database exists with another owner")

		c.JSON(http.StatusBadRequest, gin.H{"error": "Database exists with another owner"})
		return
	}

	// The database not exists
	_, err = q.IsUserExists(c, pgtype.Text{String: request.Owner, Valid: true})
	if err != nil {
		if err != pgx.ErrNoRows {
			log.Error().
				Err(err).
				Msg("Failed to check if user exists")
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "Failed to check if user exists"})
			return
		}
		// Create User here
		// TODO: Ensure request.User is valid
		// TODO: Ensure request.Password is not empty
		_, err := request.Conn.Exec(c, fmt.Sprintf("CREATE USER %s WITH PASSWORD %s",
			pq.QuoteIdentifier(request.Owner),
			pq.QuoteLiteral(request.Password)))
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed to create user")
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "Failed to create user"})
			return
		}
	}

	log.Log().Msg("User created successfully")
	// Create Database here
	_, err = request.Conn.Exec(c, fmt.Sprintf("CREATE DATABASE %s OWNER %s",
		pq.QuoteIdentifier(request.Name),
		pq.QuoteIdentifier(request.Owner)))
	if err != nil {
		log.Error().
			Err(err).
			Msg("Failed to create database")
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Failed to create database"})
		return
	}
	log.Log().Msg("Database created successfully")

	c.JSON(http.StatusCreated, gin.H{"message": "Database created successfully"})
}
