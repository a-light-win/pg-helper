package handler

import "context"

type Server interface {
	Run()
	Shutdown(ctx context.Context) error
}
