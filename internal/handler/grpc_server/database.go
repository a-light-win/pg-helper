package grpc_server

import (
	"sync"

	"github.com/a-light-win/pg-helper/pkg/proto"
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

// Return true if the database status is changed
func (d *Database) Update(db *proto.Database) bool {
	if db == nil {
		return false
	}

	d.Lock.Lock()
	defer d.Lock.Unlock()

	if d.Database != nil && db.UpdatedAt == d.UpdatedAt {
		log.Debug().Str("DbName", db.Name).
			Str("OldStage", d.Stage.String()).
			Str("OldStatus", d.Status.String()).
			Interface("OldUpdatedAt", d.UpdatedAt).
			Str("Stage", db.Stage.String()).
			Str("Status", db.Status.String()).
			Interface("UpdatedAt", db.UpdatedAt).
			Msg("database status not changed")

		return false
	}

	log.Info().Str("DbName", db.Name).
		Str("OldStage", d.Stage.String()).
		Str("OldStatus", d.Status.String()).
		Str("Stage", db.Stage.String()).
		Str("Status", db.Status.String()).
		Msg("database status changed")

	changed := d.Database == nil || d.Stage != db.Stage || d.Status != db.Status

	d.Database = db

	return changed
}

func (d *Database) Ready() bool {
	return d.Database != nil &&
		d.Stage == proto.DbStage_Ready &&
		d.Status == proto.DbStatus_Done
}

func (d *Database) IsFailed() bool {
	return d.Database != nil &&
		(d.Status == proto.DbStatus_Failed ||
			d.Status == proto.DbStatus_Cancelled ||
			d.Status == proto.DbStatus_Expired)
}
