package grpcServerApi

import (
	"time"
)

// TODO: move from this package

type InstanceFilter struct {
	// The Instance InstanceName
	InstanceName string `form:"instance_name" json:"instance_name" binding:"max=63,iname"`
	// The Postgres major version
	Version int32 `form:"version" json:"version" binding:"pg_ver"`
	// The database name
	Name string `form:"name" json:"name" binding:"required,max=63,id"`
	// Database must be in the instance
	MustExist bool `form:"must_exist" json:"must_exist" binding:""`
}

type DbRequest struct {
	InstanceFilter
}

type CreateDbRequest struct {
	InstanceFilter
	Owner       string `json:"owner" binding:"max=63,id"`
	Password    string `json:"password" binding:"required,min=8,max=256"`
	Reason      string `json:"reason" binding:"max=1024"`
	MigrateFrom string `json:"migrate_from" binding:"max=63,iname"`
}

type MigrateOutDbRequest struct {
	Name         string    `json:"name" binding:"required,max=63,id"`
	InstanceName string    `json:"instance_name" binding:"required,max=63,iname"`
	Reason       string    `json:"reason" binding:"max=1024"`
	MigrateTo    string    `json:"migrate_to" binding:"required,max=63,iname"`
	ExpireAt     time.Time `json:"-"`
}

type DbStatusResponse struct {
	Name      string    `json:"name"`
	Stage     string    `json:"stage"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
	ErrorMsg  string    `json:"error_msg"`

	InstanceName string `json:"instance_name"`
	Version      int32  `json:"version"`
}

func (status *DbStatusResponse) IsFailed() bool {
	return status.Status == "Failed" || status.Status == "Expired" || status.Status == "Cancelled"
}

func (status *DbStatusResponse) IsReady(name string, instanceName string) bool {
	return status.Stage == "Ready" &&
		status.Status == "Done" &&
		status.Name == name &&
		status.InstanceName == instanceName
}

func (status *DbStatusResponse) IsMigrateOutReady(name string, instanceName string) bool {
	return status.Stage == "Idle" &&
		status.Status == "Done" &&
		status.Name == name &&
		status.InstanceName == instanceName
}

type DbManager interface {
	GetDbStatus(request *DbRequest) (*DbStatusResponse, error)
	CreateDb(request *CreateDbRequest) error

	SubscribeDbStatus
	SubscribeInstanceStatus
}

type DbReadyWaiter interface {
	WaitReady(instName string, dbName string, timeout time.Duration) bool
}
