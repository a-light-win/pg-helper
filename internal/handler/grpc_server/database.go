package grpc_server

import (
	"context"
	"sync"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/rs/zerolog/log"
)

type Database struct {
	*proto.Database

	Lock sync.Mutex
	Cond *sync.Cond
}

func NewDatabase() *Database {
	newDb := &Database{Database: &proto.Database{}}
	newDb.Cond = sync.NewCond(&newDb.Lock)
	return newDb
}

func (d *Database) NotifyChanged() {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	d.notifyChanged()
}

func (d *Database) notifyChanged() {
	d.Cond.Broadcast()
}

func (d *Database) Update(db *proto.Database) {
	if db == nil {
		return
	}

	d.Lock.Lock()
	defer d.Lock.Unlock()

	if d.Database != nil && db.UpdatedAt == d.UpdatedAt {
		log.Warn().Str("DbName", db.Name).
			Str("OldStage", d.Stage.String()).
			Str("OldStatus", d.Status.String()).
			Interface("OldUpdatedAt", d.UpdatedAt).
			Str("Stage", db.Stage.String()).
			Str("Status", db.Status.String()).
			Interface("UpdatedAt", db.UpdatedAt).
			Msg("database status not changed")

		return
	}

	log.Log().Str("DbName", db.Name).
		Str("OldStage", d.Stage.String()).
		Str("OldStatus", d.Status.String()).
		Str("Stage", db.Stage.String()).
		Str("Status", db.Status.String()).
		Msg("database status changed")

	d.Database = db
	d.notifyChanged()
}

func (d *Database) retry(sender proto.DbTaskSvc_RegisterServer) {
	// TODO: retry logic
	// If backup failed, resend create database task later
	// If restore failed, reset the database and then resend create database task later
}

func (d *Database) WaitReady(timeoutCtx context.Context) bool {
	d.Lock.Lock()
	defer d.Lock.Unlock()

	if timeoutCtx != nil {
		go func() {
			<-timeoutCtx.Done()
			d.NotifyChanged()
		}()
	}

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

func (d *Database) IsProcessing() bool {
	return d.Database != nil && (d.Stage == proto.DbStage_Creating || d.Stage == proto.DbStage_Backuping || d.Stage == proto.DbStage_Restoring)
}
