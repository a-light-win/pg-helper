package handler

import "github.com/jackc/pgx/v5/pgxpool"

type Handler struct {
	DbPool *pgxpool.Pool
}

func New(dbPool *pgxpool.Pool) *Handler {
	return &Handler{DbPool: dbPool}
}
