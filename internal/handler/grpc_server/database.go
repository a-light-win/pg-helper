package grpc_server

import (
	"context"
	"sync"

	"github.com/a-light-win/pg-helper/api/proto"
)

type Database struct {
	*proto.Database

	Lock sync.Mutex
	Cond *sync.Cond
}

func NewDatabase() *Database {
	newDb := &Database{}
	newDb.Cond = sync.NewCond(&newDb.Lock)
	return newDb
}

func (d *Database) NotifyChanged() {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	d.Cond.Broadcast()
}

func (d *Database) Update(db *proto.Database) {
	if db == nil {
		return
	}

	d.Lock.Lock()
	defer d.Lock.Unlock()

	if d.Database != nil && db.UpdatedAt == d.UpdatedAt {
		return
	}

	d.Database = db
	d.NotifyChanged()
}

func (d *Database) retry(sender proto.DbTaskSvc_RegisterServer) {
	// TODO: retry logic
	// If backup failed, resend create database task later
	// If restore failed, reset the database and then resend create database task later
}

func (d *Database) ProtoDatabase() *proto.Database {
	return d.Database
}

func (d *Database) WaitReady(timeoutCtx context.Context) bool {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	for {
		if d.Ready() {
			return true
		}
		if timeoutCtx != nil {
			select {
			case <-timeoutCtx.Done():
				return false
			default:
			}
		}
		d.Cond.Wait()
	}
}

func (d *Database) Ready() bool {
	return d.Database != nil &&
		d.Stage == proto.DbStage_Running &&
		d.Status == proto.DbStatus_Done
}
