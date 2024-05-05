package grpc_server

import (
	"context"
	"sync"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/rs/zerolog"
)

type DbInstance struct {
	Name      string
	PgVersion int32

	Databases map[string]*Database
	// Protects Databases
	dbLock sync.Mutex

	DbTaskChan    chan *proto.DbTask
	nonSentDbTask *proto.DbTask

	logger *zerolog.Logger
}

func NewDbInstance(name string, pgVersion int32, logger *zerolog.Logger) *DbInstance {
	return &DbInstance{
		Name:       name,
		PgVersion:  pgVersion,
		Databases:  make(map[string]*Database),
		DbTaskChan: make(chan *proto.DbTask),

		logger: logger,
	}
}

func (a *DbInstance) UpdateDatabases(ctx context.Context, databases []*proto.Database) {
	a.dbLock.Lock()
	defer a.dbLock.Unlock()

	for _, db := range databases {
		oldDb := a.mustGetDb(ctx, db.Name)
		oldDb.Update(db)
	}
}

func (a *DbInstance) UpdateDatabase(ctx context.Context, db *proto.Database) {
	oldDb := a.MustGetDb(ctx, db.Name)
	oldDb.Update(db)
}

func (a *DbInstance) GetDb(name string) *Database {
	a.dbLock.Lock()
	defer a.dbLock.Unlock()

	if db, ok := a.Databases[name]; ok {
		return db
	}
	return nil
}

func (a *DbInstance) MustGetDb(ctx context.Context, name string) *Database {
	a.dbLock.Lock()
	defer a.dbLock.Unlock()

	return a.mustGetDb(ctx, name)
}

func (a *DbInstance) mustGetDb(ctx context.Context, name string) *Database {
	db, ok := a.Databases[name]

	if !ok {
		db = NewDatabase(ctx)
		a.Databases[name] = db
	}

	return db
}

func (a *DbInstance) ServeDbTask(s proto.DbTaskSvc_RegisterServer) {
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
		case task := <-a.DbTaskChan:
			if err := s.Send(task); err != nil {
				a.nonSentDbTask = task
				return
			}
		}
	}
}

func (a *DbInstance) Send(task *proto.DbTask) {
	a.DbTaskChan <- task
}
