package grpc_server

import (
	"errors"
	"sync"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/google/uuid"
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

func (a *DbInstance) UpdateDatabases(databases []*proto.Database) {
	a.dbLock.Lock()
	defer a.dbLock.Unlock()

	for _, db := range databases {

		a.logger.Debug().
			Str("DbName", db.Name).
			Str("Stage", db.Stage.String()).
			Str("Status", db.Status.String()).
			Msg("Init database")

		oldDb := a.mustGetDb(db.Name)
		oldDb.Update(db)
	}
}

func (a *DbInstance) UpdateDatabase(db *proto.Database) {
	a.logger.Debug().
		Str("DbName", db.Name).
		Str("Stage", db.Stage.String()).
		Str("Status", db.Status.String()).
		Msg("Update database")

	oldDb := a.MustGetDb(db.Name)
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

func (a *DbInstance) MustGetDb(name string) *Database {
	a.dbLock.Lock()
	defer a.dbLock.Unlock()

	return a.mustGetDb(name)
}

func (a *DbInstance) mustGetDb(name string) *Database {
	db, ok := a.Databases[name]

	if !ok {
		db = NewDatabase()
		a.Databases[name] = db
	}

	return db
}

func (a *DbInstance) ServeDbTask(s proto.DbTaskSvc_RegisterServer) {
	if a.nonSentDbTask != nil {
		if err := s.Send(a.nonSentDbTask); err != nil {
			a.logger.Error().Err(err).
				Str("request_id", a.nonSentDbTask.RequestId).
				Msg("Resent non-sent db task failed")
			return
		}
		a.logger.Debug().
			Str("request_id", a.nonSentDbTask.RequestId).
			Msg("Resend non-sent db task success")
		a.nonSentDbTask = nil
	}

	for {
		select {
		case <-s.Context().Done():
			return
		case task := <-a.DbTaskChan:
			if err := s.Send(task); err != nil {
				a.logger.Error().Err(err).
					Str("request_id", task.RequestId).
					Msg("Sent db task failed")
				a.nonSentDbTask = task
				return
			}
			a.logger.Debug().
				Str("request_id", task.RequestId).
				Msg("Sent db task success")
		}
	}
}

func (a *DbInstance) Send(task *proto.DbTask) {
	a.DbTaskChan <- task
}

func (a *DbInstance) IsDbReady(dbName string) bool {
	db := a.GetDb(dbName)
	return db != nil && db.Ready()
}

func (a *DbInstance) CreateDb(vo *handler.CreateDbRequest) (*Database, error) {
	a.dbLock.Lock()
	defer a.dbLock.Unlock()

	db := a.mustGetDb(vo.DbName)
	if db.IsProcessing() {
		return db, nil
	}
	if db.Stage == proto.DbStage_Dropping {
		return nil, errors.New("database is dropping")
	}
	if db.Stage == proto.DbStage_MigrateOut {
		// TODO: rollback to previous version here?
		return nil, errors.New("database is migrating out")
	}

	task := &proto.DbTask{
		RequestId: uuid.New().String(),
		Task: &proto.DbTask_CreateDatabase{
			CreateDatabase: &proto.CreateDatabaseTask{
				Name:        vo.DbName,
				Reason:      vo.Reason,
				Owner:       vo.DbOwner,
				Password:    vo.DbPassword,
				MigrateFrom: vo.MigrateFrom,
			},
		},
	}
	a.logger.Debug().Str("DbName", vo.DbName).Msg("Task to create database")
	a.Send(task)

	return db, nil
}
