package grpc_handler

import (
	"sync"

	"github.com/a-light-win/pg-helper/api/proto"
)

type AgentData struct {
	ID        string
	PgVersion int32

	Databases  map[string]*Database
	TaskSender proto.DbTaskSvc_RegisterServer

	// Protects Databases
	Lock sync.Mutex
}

func (a *AgentData) UpdateDatabase(db *proto.Database) {
	oldDb := a.MustGetDb(db.Name)
	oldDb.Update(db)
}

func (a *AgentData) MustGetDb(name string) *Database {
	a.Lock.Lock()
	defer a.Lock.Unlock()

	db, ok := a.Databases[name]
	if !ok {
		db = NewDatabase()
		a.Databases[name] = db
	}
	return db
}
