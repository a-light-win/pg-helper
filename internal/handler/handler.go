package handler

import "context"

type Server interface {
	Init(config any) error
	Run()
	Shutdown(ctx context.Context) error
}
