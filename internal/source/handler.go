package source

import (
	"fmt"
	"sync"
	"time"

	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/internal/interface/grpc_server"
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/a-light-win/pg-helper/pkg/utils/logger"
	"github.com/a-light-win/pg-helper/pkg/validate"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type SourceHandler struct {
	Config *config.SourceConfig

	Databases      map[string]*DatabaseSource
	databasesMutex sync.Mutex

	Instances      map[string]bool
	instancesMutex sync.Mutex

	cronProducer   server.Producer
	sourceProducer server.Producer
	dbManager      grpc_server.DbManager

	validator *validator.Validate
}

type ParentSourceHandler interface {
	AddDatabaseSource(source *DatabaseSource) error
	MarkDatabaseSourceIdle(name string) error
}

func NewSourceHandler(config *config.SourceConfig) *SourceHandler {
	return &SourceHandler{
		Config:    config,
		Databases: make(map[string]*DatabaseSource),
		Instances: make(map[string]bool),
		validator: validate.New(),
	}
}

func (h *SourceHandler) Init(setter server.GlobalSetter) error {
	return nil
}

func (h *SourceHandler) PostInit(getter server.GlobalGetter) error {
	h.cronProducer = getter.Get(constants.ServerKeyCronProducer).(server.Producer)
	h.sourceProducer = getter.Get(constants.ServerKeySourceProducer).(server.Producer)
	h.dbManager = getter.Get(constants.ServerKeyDbManager).(grpc_server.DbManager)

	h.dbManager.SubscribeDbStatus(h.OnDbStatusChanged)
	h.dbManager.SubscribeInstanceStatus(h.OnInstanceStatusChanged)
	return nil
}

func (h *SourceHandler) Handle(msg server.NamedElement) error {
	source := msg.(*DatabaseSource)
	source.LastScheduledAt = time.Now()

	if h.syncDatabaseSource(source) {
		return nil
	}

	source.State = SourceStateProcessing

	if source.ExpectState == SourceStateIdle {
		// TODO: mark database as idle and drop it later (a month or more later?)
		// we do not remove it immediately because we want to keep the data for a while
		// in case we need to rollback.
		return nil
	}

	dbPassword, err := source.PasswordContent()
	if err != nil {
		log.Warn().Err(err).
			Str("DbName", source.Name).
			Msg("Can not get password of the database owner")
		return logger.NewAlreadyLoggedError(err, zerolog.WarnLevel)
	}

	request := &grpc_server.CreateDbRequest{
		InstanceFilter: grpc_server.InstanceFilter{
			Name:   source.InstanceName,
			DbName: source.Name,
		},
		DbOwner:     source.Owner,
		DbPassword:  dbPassword,
		MigrateFrom: source.MigrateFrom,
		Reason:      fmt.Sprintf("Create database %s from source %s", source.Name, source.Type),
	}

	if _, err := h.dbManager.CreateDb(request, false); err != nil {
		log.Warn().Err(err).
			Str("DbName", source.Name).
			Msg("Failed to create database")

		source.LastErrorMsg = err.Error()
		source.UpdatedAt = time.Now()
		if err == grpc_server.ErrInstanceOffline {
			source.State = SourceStatePending
		} else {
			source.State = SourceStateFailed
			h.retryNextTime(source)
		}
		return logger.NewAlreadyLoggedError(err, zerolog.WarnLevel)
	}

	return nil
}

func (h *SourceHandler) AddDatabaseSource(source *DatabaseSource) error {
	if err := h.validator.Struct(source); err != nil {
		return err
	}

	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()
	if oldSource, ok := h.Databases[source.Name]; ok {
		if !oldSource.IsConfigChanged(source) {
			log.Debug().Str("source", source.Name).Msg("Source not changed, skip")
			return nil
		}
		log.Debug().Str("Name", source.Name).Msg("source changed")
	} else {
		log.Debug().Str("Name", source.Name).Msg("source added")
	}

	if source.ExpectState == "" {
		source.ExpectState = SourceStateReady
	}
	h.Databases[source.Name] = source

	go h.sourceProducer.Send(source)

	return nil
}

func (h *SourceHandler) MarkDatabaseSourceIdle(name string) error {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()

	if source, ok := h.Databases[name]; ok {
		if source.ExpectState != SourceStateIdle {
			source.ExpectState = SourceStateIdle
			source.State = SourceStateScheduling
			source.NextScheduleAt = time.Now().Add(h.Config.DeleyDelete)
			h.cronProducer.Send(&server.CronElement{
				TriggerAt: source.NextScheduleAt,
				HandleFunc: func(triggerAt time.Time) {
					h.idleDatabaseSource(name, triggerAt)
				},
			})
		}
	}
	return nil
}

func (h *SourceHandler) idleDatabaseSource(name string, triggerAt time.Time) {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()
	if source, ok := h.Databases[name]; ok {
		if source.ExpectState == SourceStateIdle && source.NextScheduleAt.Equal(triggerAt) {
			go h.sourceProducer.Send(source)
		}
	}
}

func (h *SourceHandler) OnDbStatusChanged(dbStatus *grpc_server.DbStatusResponse) bool {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()

	if source, ok := h.Databases[dbStatus.Name]; ok {
		h.updateDbStatus(source, dbStatus)
	}
	return true
}

func (h *SourceHandler) updateDbStatus(source *DatabaseSource, dbStatus *grpc_server.DbStatusResponse) {
	if source.UpdateState(dbStatus) {
		if source.State == SourceStateFailed {
			h.retryNextTime(source)
			return
		}
		if source.State == SourceStateDropped {
			go h.RemoveDatabaseSource(source)
			return
		}
	}
}

func (h *SourceHandler) RemoveDatabaseSource(source *DatabaseSource) {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()

	if source.State == SourceStateDropped {
		delete(h.Databases, source.Name)
	}
}

// return true if the source is synced else false
func (h *SourceHandler) syncDatabaseSource(source *DatabaseSource) bool {
	dbRequest := &grpc_server.DbRequest{
		InstanceFilter: grpc_server.InstanceFilter{
			Name:   source.InstanceName,
			DbName: source.Name,
		},
	}
	dbStatus, err := h.dbManager.GetDbStatus(dbRequest)
	if err != nil {
		return false
	}

	h.updateDbStatus(source, dbStatus)
	return source.Synced()
}

func (h *SourceHandler) retryNextTime(source *DatabaseSource) {
	source.State = SourceStateScheduling
	source.NextScheduleAt = time.Now().Add(source.NextRetryDelay())
	log.Debug().Int("RetryDelay", source.RetryDelay).
		Int("RetryTimes", source.RetryTimes).
		Str("DbName", source.Name).
		Interface("NextScheduleAt", source.NextScheduleAt).
		Msg("Retry next time")

	h.cronProducer.Send(&server.CronElement{
		TriggerAt: source.NextScheduleAt,
		HandleFunc: func(triggerAt time.Time) {
			h.databasesMutex.Lock()
			defer h.databasesMutex.Unlock()
			if source_, ok := h.Databases[source.Name]; ok {
				if source_.NextScheduleAt.Equal(triggerAt) {
					go h.sourceProducer.Send(source_)
				}
			}
		},
	})
}

func (h *SourceHandler) OnInstanceStatusChanged(instanceStatus *grpc_server.InstanceStatusResponse) bool {
	h.instancesMutex.Lock()
	defer h.instancesMutex.Unlock()

	if online, ok := h.Instances[instanceStatus.Name]; ok {
		if online == instanceStatus.Online {
			return true
		}
	}

	h.Instances[instanceStatus.Name] = instanceStatus.Online
	if instanceStatus.Online {
		go h.handleDatabasesOnInstanceOnline(instanceStatus)
	}
	return true
}

func (h *SourceHandler) handleDatabasesOnInstanceOnline(instance *grpc_server.InstanceStatusResponse) {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()
	for _, source := range h.Databases {
		if source.Synced() {
			continue
		}
		if source.InstanceName != instance.Name && source.MigrateFrom != instance.Name {
			continue
		}

		if database, ok := instance.Databases[source.Name]; ok {
			h.updateDbStatus(source, database)
		}

		if source.State == SourceStatePending {
			source.State = SourceStateScheduling
			source.NextScheduleAt = time.Now()
			go h.sourceProducer.Send(source)
		}
	}
}
