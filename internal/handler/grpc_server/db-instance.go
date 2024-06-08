package grpc_server

import (
	"errors"
	"sync"

	api "github.com/a-light-win/pg-helper/internal/interface/grpcServerApi"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type DbInstance struct {
	Name      string
	PgVersion int32
	Online    bool

	Databases map[string]*Database
	// Protects Databases
	dbLock sync.Mutex

	DbJobChan    chan *proto.DbJob
	nonSentDbJob *proto.DbJob

	logger *zerolog.Logger

	subscriber *DbStatusSubscriber
}

func NewDbInstance(name string, pgVersion int32, logger *zerolog.Logger, subcriber *DbStatusSubscriber) *DbInstance {
	return &DbInstance{
		Name:      name,
		PgVersion: pgVersion,
		Databases: make(map[string]*Database),
		DbJobChan: make(chan *proto.DbJob),

		logger:     logger,
		subscriber: subcriber,
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
		if oldDb.Update(db) {
			go a.subscriber.OnStatusChanged(a, oldDb)
		}
	}
}

func (a *DbInstance) UpdateDatabase(db *proto.Database) {
	a.logger.Debug().
		Str("DbName", db.Name).
		Str("Stage", db.Stage.String()).
		Str("Status", db.Status.String()).
		Msg("Update database")

	oldDb := a.MustGetDb(db.Name)
	if oldDb.Update(db) {
		go a.subscriber.OnStatusChanged(a, oldDb)
	}
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

func (a *DbInstance) ServeDbJob(s proto.DbJobSvc_RegisterServer) {
	if a.nonSentDbJob != nil {
		if err := s.Send(a.nonSentDbJob); err != nil {
			a.logger.Error().Err(err).
				Str("job_id", a.nonSentDbJob.JobId).
				Msg("Resent non-sent db job failed")
			return
		}
		a.logger.Debug().
			Str("job_id", a.nonSentDbJob.JobId).
			Msg("Resend non-sent db job success")
		a.nonSentDbJob = nil
	}

	for {
		select {
		case <-s.Context().Done():
			return
		case job := <-a.DbJobChan:
			if err := s.Send(job); err != nil {
				a.logger.Error().Err(err).
					Str("job_id", job.JobId).
					Msg("Sent db job failed")
				a.nonSentDbJob = job
				return
			}
			a.logger.Debug().
				Str("job_id", job.JobId).
				Msg("Sent db job success")
		}
	}
}

func (a *DbInstance) Send(job *proto.DbJob) {
	a.DbJobChan <- job
}

func (a *DbInstance) CreateDb(vo *api.CreateDbRequest) error {
	a.dbLock.Lock()
	defer a.dbLock.Unlock()

	db := a.mustGetDb(vo.Name)

	if db.Stage == proto.DbStage_DropDatabase {
		return errors.New("database is dropping")
	}
	if db.Stage == proto.DbStage_Idle {
		// TODO: rollback to previous version here?
		return errors.New("database is migrating out")
	}

	if db.Stage != proto.DbStage_None && !db.IsFailed() {
		return nil
	}

	job := &proto.DbJob{
		JobId: uuid.New().String(),
		Job: &proto.DbJob_CreateDatabase{
			CreateDatabase: &proto.CreateDatabaseJob{
				Name:        vo.Name,
				Reason:      vo.Reason,
				Owner:       vo.Owner,
				Password:    vo.Password,
				MigrateFrom: vo.MigrateFrom,
				BackupPath:  vo.BackupPath,
			},
		},
	}
	a.logger.Debug().Str("DbName", vo.Name).Msg("Job to create database")
	a.Send(job)

	return nil
}

// Return true if send the migrateOut job
func (a *DbInstance) MigrateOut(request *api.MigrateOutDbRequest, callback func() error) error {
	a.dbLock.Lock()

	db, ok := a.Databases[request.Name]
	if !ok || db.IsReadyToMigrate() {
		a.dbLock.Unlock()
		return callback()
	}

	defer a.dbLock.Unlock()

	a.subscriber.Subscribe(func(dbStatus *api.DbStatusResponse) bool {
		if dbStatus.IsMigrateOutReady(request.Name, a.Name) {
			go callback()
			return api.StopSubscribe
		}
		return api.ContinueSubscribe
	})

	job := &proto.DbJob{
		Job: &proto.DbJob_MigrateOutDatabase{
			MigrateOutDatabase: &proto.MigrateOutDatabaseJob{
				Name:      request.Name,
				Reason:    request.Reason,
				MigrateTo: request.MigrateTo,
			},
		},
	}
	a.Send(job)
	return nil
}

func (a *DbInstance) StatusResponse() *api.InstanceStatusResponse {
	a.dbLock.Lock()
	defer a.dbLock.Unlock()

	databases := make(map[string]*api.DbStatusResponse)
	if a.Online {
		for _, db := range a.Databases {
			dbStatus := db.StatusResponse()
			dbStatus.InstanceName = a.Name
			dbStatus.Version = a.PgVersion
			databases[db.Name] = dbStatus
		}
	}

	return &api.InstanceStatusResponse{
		Name:    a.Name,
		Version: a.PgVersion,
		Online:  a.Online,

		Databases: databases,
	}
}
