package grpc_handler

import (
	"github.com/a-light-win/pg-helper/api/proto"
)

type GrpcHandler interface {
	Validate() error
	Handle() error
	OnError(err error)
}

func New(task *proto.DbTask) GrpcHandler {
	switch task.Task.(type) {
	case *proto.DbTask_CreateDatabase:
		return NewCreateDatabaseHandler(task)
	}
	return nil
}
