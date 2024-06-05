package source

import (
	"fmt"
	"sync"
	"time"

	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/internal/interface/grpcServerApi"
	"github.com/a-light-win/pg-helper/internal/interface/sourceApi"
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/a-light-win/pg-helper/pkg/utils/logger"
	"github.com/a-light-win/pg-helper/pkg/validate"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type BaseSourceHandler struct {
	Config *config.SourceConfig

	Databases      map[string]*sourceApi.DatabaseSource
	databasesMutex sync.Mutex

	Instances      map[string]bool
	instancesMutex sync.Mutex

	cronProducer   server.Producer
	sourceProducer server.Producer
	dbManager      grpcServerApi.DbManager

	validator *validator.Validate
}

func NewSourceHandler(config *config.SourceConfig) *BaseSourceHandler {
	return &BaseSourceHandler{
		Config:    config,
		Databases: make(map[string]*sourceApi.DatabaseSource),
		Instances: make(map[string]bool),
		validator: validate.New(),
	}
}

func (h *BaseSourceHandler) Init(setter server.GlobalSetter) error {
	return nil
}

func (h *BaseSourceHandler) PostInit(getter server.GlobalGetter) error {
	h.cronProducer = getter.Get(constants.ServerKeyCronProducer).(server.Producer)
	h.sourceProducer = getter.Get(constants.ServerKeySourceProducer).(server.Producer)
	h.dbManager = getter.Get(constants.ServerKeyDbManager).(grpcServerApi.DbManager)

	h.dbManager.SubscribeDbStatus(h.OnDbStatusChanged)
	h.dbManager.SubscribeInstanceStatus(h.OnInstanceStatusChanged)
	return nil
}

func (h *BaseSourceHandler) Handle(msg server.NamedElement) error {
	source := msg.(*sourceApi.DatabaseSource)
	source.LastScheduledAt = time.Now()

	if h.syncDatabaseSource(source) {
		return nil
	}

	source.State = sourceApi.SourceStateProcessing

	if source.ExpectState == sourceApi.SourceStateIdle {
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

	request := &grpcServerApi.CreateDbRequest{
		InstanceFilter: grpcServerApi.InstanceFilter{
			InstanceName: source.InstanceName,
			Name:         source.Name,
		},
		Owner:       source.Owner,
		Password:    dbPassword,
		MigrateFrom: source.MigrateFrom,
		Reason:      fmt.Sprintf("Create database %s from source %s", source.Name, source.Type),
	}

	if _, err := h.dbManager.CreateDb(request, false); err != nil {
		log.Warn().Err(err).
			Str("DbName", source.Name).
			Msg("Failed to create database")

		source.LastErrorMsg = err.Error()
		source.UpdatedAt = time.Now()
		if err == grpcServerApi.ErrInstanceOffline {
			source.State = sourceApi.SourceStatePending
		} else {
			source.State = sourceApi.SourceStateFailed
			h.retryNextTime(source)
		}
		return logger.NewAlreadyLoggedError(err, zerolog.WarnLevel)
	}

	return nil
}

func (h *BaseSourceHandler) AddDatabaseSource(source *sourceApi.DatabaseSource) error {
	if err := h.validator.Struct(source); err != nil {
		return err
	}

	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()
	if oldSource, ok := h.Databases[source.Name]; ok {
		if !oldSource.IsConfigChanged(source.DatabaseRequest) {
			log.Debug().Str("source", source.Name).Msg("Source not changed, skip")
			return nil
		}
		log.Debug().Str("Name", source.Name).Msg("source changed")
	} else {
		log.Debug().Str("Name", source.Name).Msg("source added")
	}

	if source.ExpectState == "" {
		source.ExpectState = sourceApi.SourceStateReady
	}
	h.Databases[source.Name] = source

	go h.sourceProducer.Send(source)

	return nil
}

func (h *BaseSourceHandler) MarkDatabaseSourceIdle(name string) error {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()

	if source, ok := h.Databases[name]; ok {
		if source.ExpectState != sourceApi.SourceStateIdle {
			source.ExpectState = sourceApi.SourceStateIdle
			source.State = sourceApi.SourceStateScheduling
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

func (h *BaseSourceHandler) idleDatabaseSource(name string, triggerAt time.Time) {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()
	if source, ok := h.Databases[name]; ok {
		if source.ExpectState == sourceApi.SourceStateIdle && source.NextScheduleAt.Equal(triggerAt) {
			go h.sourceProducer.Send(source)
		}
	}
}

func (h *BaseSourceHandler) OnDbStatusChanged(dbStatus *grpcServerApi.DbStatusResponse) bool {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()

	if source, ok := h.Databases[dbStatus.Name]; ok {
		h.updateDbStatus(source, dbStatus)
	}
	return true
}

func (h *BaseSourceHandler) updateDbStatus(source *sourceApi.DatabaseSource, dbStatus *grpcServerApi.DbStatusResponse) {
	if source.UpdateState(dbStatus) {
		if source.State == sourceApi.SourceStateFailed {
			h.retryNextTime(source)
			return
		}
		if source.State == sourceApi.SourceStateDropped {
			go h.RemoveDatabaseSource(source)
			return
		}
	}
}

func (h *BaseSourceHandler) RemoveDatabaseSource(source *sourceApi.DatabaseSource) {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()

	if source.State == sourceApi.SourceStateDropped {
		delete(h.Databases, source.Name)
	}
}

// return true if the source is synced else false
func (h *BaseSourceHandler) syncDatabaseSource(source *sourceApi.DatabaseSource) bool {
	dbRequest := &grpcServerApi.DbRequest{
		InstanceFilter: grpcServerApi.InstanceFilter{
			InstanceName: source.InstanceName,
			Name:         source.Name,
		},
	}
	dbStatus, err := h.dbManager.GetDbStatus(dbRequest)
	if err != nil {
		return false
	}

	h.updateDbStatus(source, dbStatus)
	return source.Synced()
}

func (h *BaseSourceHandler) retryNextTime(source *sourceApi.DatabaseSource) {
	source.State = sourceApi.SourceStateScheduling
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

func (h *BaseSourceHandler) OnInstanceStatusChanged(instanceStatus *grpcServerApi.InstanceStatusResponse) bool {
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

func (h *BaseSourceHandler) handleDatabasesOnInstanceOnline(instance *grpcServerApi.InstanceStatusResponse) {
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

		if source.State == sourceApi.SourceStatePending {
			source.State = sourceApi.SourceStateScheduling
			source.NextScheduleAt = time.Now()
			go h.sourceProducer.Send(source)
		}
	}
}

func (h *BaseSourceHandler) IsReady(dbName, instanceName string) bool {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()

	if source, ok := h.Databases[dbName]; ok {
		return source.InstanceName == instanceName && source.State == sourceApi.SourceStateReady
	}
	return false
}

func (h *BaseSourceHandler) GetSource(name string) *sourceApi.DatabaseSource {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()
	if source, ok := h.Databases[name]; ok {
		return source
	}
	return nil
}
