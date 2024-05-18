package handler

import (
	"github.com/a-light-win/pg-helper/api/proto"
)

type InstanceFilter struct {
	// The Instance Name
	Name string `binding:"max=63,id"`
	// The Postgres major version
	Version int32 `binding:"pg_ver"`
	// The database name
	DbName string `binding:"required,max=63,id"`
	// Database must be in the instance
	MustExist bool `binding:""`
}

type DbVO struct {
	InstanceFilter
}

type CreateDbVO struct {
	InstanceFilter
	DbOwner     string `binding:"max=63,id"`
	DbPassword  string `binding:"required,min=8,max=256"`
	Reason      string `binding:"max=1024"`
	MigrateFrom string `binding:"max=63,id"`
}

type DbManager interface {
	IsDbReady(vo *DbVO) bool
	CreateDb(vo *CreateDbVO) (*proto.Database, error)
}
