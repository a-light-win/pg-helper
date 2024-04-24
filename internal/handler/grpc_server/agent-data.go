package grpc_server

import (
	"sync"

	"github.com/a-light-win/pg-helper/api/proto"
)

type AgentData struct {
	ID        string
	PgVersion int32

	Databases map[string]*Database
	// Protects Databases
	dbLock sync.Mutex

	DbTaskChan    chan *proto.DbTask
	nonSentDbTask *proto.DbTask
}

func NewAgentData(agentId string, pgVersion int32) *AgentData {
	return &AgentData{
		ID:         agentId,
		PgVersion:  pgVersion,
		Databases:  make(map[string]*Database),
		DbTaskChan: make(chan *proto.DbTask),
	}
}

func (a *AgentData) UpdateDatabases(databases []*proto.Database) {
	a.dbLock.Lock()
	defer a.dbLock.Unlock()

	for _, db := range databases {
		oldDb := a.mustGetDb(db.Name)
		oldDb.Update(db)
	}
}

func (a *AgentData) UpdateDatabase(db *proto.Database) {
	oldDb := a.MustGetDb(db.Name)
	oldDb.Update(db)
}

func (a *AgentData) MustGetDb(name string) *Database {
	a.dbLock.Lock()
	defer a.dbLock.Unlock()

	return a.mustGetDb(name)
}

func (a *AgentData) mustGetDb(name string) *Database {
	db, ok := a.Databases[name]

	if !ok {
		db = NewDatabase()
		a.Databases[name] = db
	}

	return db
}

func (a *AgentData) ServeDbTask(s proto.DbTaskSvc_RegisterServer) {
	if a.nonSentDbTask != nil {
		if err := s.Send(a.nonSentDbTask); err != nil {
			return
		}
		a.nonSentDbTask = nil
	}

	for {
		select {
		case <-s.Context().Done():
			return
		case <-gd_.QuitCtx.Done():
			return
		case task := <-a.DbTaskChan:
			if err := s.Send(task); err != nil {
				a.nonSentDbTask = task
				return
			}
		}
	}
}
