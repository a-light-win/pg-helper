package grpc_handler

import (
	"context"
	"sync"

	"github.com/a-light-win/pg-helper/api/proto"
)

type Database struct {
	*proto.Database

	Lock sync.Mutex
	Cond *sync.Cond

	StopCtx context.Context
	Stop    context.CancelFunc
}

func NewDatabase() *Database {
	newDb := &Database{}
	newDb.Cond = sync.NewCond(&newDb.Lock)
	newDb.StopCtx, newDb.Stop = context.WithCancel(gd_.QuitCtx)
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

func (d *Database) RetryUntilReady(sender proto.DbTaskSvc_RegisterServer) {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	for {
		select {
		case <-d.StopCtx.Done():
			return
		default:
			if d.Ready() {
				return
			}
			d.Cond.Wait()
		}

		if d.Status == proto.DbStatus_Failed {
			d.retry(sender)
		}
	}
}

func (d *Database) retry(sender proto.DbTaskSvc_RegisterServer) {
	// TODO: retry logic
	// If backup failed, resend create database task later
	// If restore failed, reset the database and then resend create database task later
}

func (d *Database) WaitReady() bool {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	for {
		select {
		case <-d.StopCtx.Done():
			return false
		default:
			if d.Ready() {
				return true
			}
			d.Cond.Wait()
		}
	}
}

func (d *Database) Ready() bool {
	return d.Database != nil &&
		d.Stage == proto.DbStage_Running &&
		d.Status == proto.DbStatus_Done
}
