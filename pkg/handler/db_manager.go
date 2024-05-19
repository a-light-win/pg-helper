package handler

import (
	"github.com/a-light-win/pg-helper/api/proto"
)

type InstanceFilter struct {
	// The Instance Name
	Name string `form:"name" json:"name" binding:"max=63"`
	// The Postgres major version
	Version int32 `form:"version" json:"version" binding:"pg_ver"`
	// The database name
	DbName string `form:"db_name" json:"db_name" binding:"required,max=63,id"`
	// Database must be in the instance
	MustExist bool `form:"must_exist" json:"must_exist" binding:""`
}

type DbVO struct {
	InstanceFilter
}

type CreateDbVO struct {
	InstanceFilter
	DbOwner     string `json:"db_owner" binding:"max=63,id"`
	DbPassword  string `json:"db_password" binding:"required,min=8,max=256"`
	Reason      string `json:"reason" binding:"max=1024"`
	MigrateFrom string `json:"migrate_from" binding:"max=63"`
}

type DbManager interface {
	IsDbReady(vo *DbVO) bool
	CreateDb(vo *CreateDbVO) (*proto.Database, error)
}
