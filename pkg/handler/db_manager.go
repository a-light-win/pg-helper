package handler

import (
	"time"
)

// TODO: move from this package

type InstanceFilter struct {
	// The Instance Name
	Name string `form:"name" json:"name" binding:"max=63,iname"`
	// The Postgres major version
	Version int32 `form:"version" json:"version" binding:"pg_ver"`
	// The database name
	DbName string `form:"db_name" json:"db_name" binding:"required,max=63,id"`
	// Database must be in the instance
	MustExist bool `form:"must_exist" json:"must_exist" binding:""`
}

type DbRequest struct {
	InstanceFilter
}

type CreateDbRequest struct {
	InstanceFilter
	DbOwner     string `json:"db_owner" binding:"max=63,id"`
	DbPassword  string `json:"db_password" binding:"required,min=8,max=256"`
	Reason      string `json:"reason" binding:"max=1024"`
	MigrateFrom string `json:"migrate_from" binding:"max=63,iname"`
}

type DbStatusResponse struct {
	Name      string    `json:"name"`
	Stage     string    `json:"stage"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`

	InstanceName string `json:"instance_name"`
	Version      int32  `json:"version"`
}

// SubscribeDbStatusFunc is a callback function to be called when the status of the database changes
// The function should return true if it wants to continue receiving notifications
// and false if it wants to unsubscribe
type SubscribeDbStatusFunc func(*DbStatusResponse) bool

type DbManager interface {
	IsDbReady(request *DbRequest) bool
	CreateDb(request *CreateDbRequest, waitReady bool) (*DbStatusResponse, error)

	// Subscribe to database status changes, the callback will be called when the status changes
	// The callback should return true if it wants to continue receiving notifications
	// and false if it wants to unsubscribe
	//
	// This function will report on following stage changed:
	// - Ready
	// - Idle
	// - DropCompleted
	SubscribeDbStatus(callback SubscribeDbStatusFunc)
}
