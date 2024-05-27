package source

import (
	"fmt"
	"sync"
	"time"

	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/a-light-win/pg-helper/pkg/validate"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type SourceHandler struct {
	Config         *config.SourceConfig
	Databases      map[string]*DatabaseSource
	databasesMutex sync.Mutex

	cronProducer   server.Producer
	sourceProducer server.Producer
	dbManager      handler.DbManager

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
		validator: validate.New(),
	}
}

func (h *SourceHandler) Init(setter server.GlobalSetter) error {
	return nil
}

func (h *SourceHandler) PostInit(getter server.GlobalGetter) error {
	h.cronProducer = getter.Get(constants.ServerKeyCronProducer).(server.Producer)
	h.sourceProducer = getter.Get(constants.ServerKeySourceProducer).(server.Producer)
	h.dbManager = getter.Get(constants.ServerKeyDbManager).(handler.DbManager)

	h.dbManager.SubscribeDbStatus(h.OnDbStatusChanged)
	return nil
}

func (h *SourceHandler) Handle(msg server.NamedElement) error {
	source := msg.(*DatabaseSource)
	source.LastScheduledTime = time.Now()

	if h.syncDatabaseSource(source) {
		source.ResetRetryDelay()
		return nil
	}

	if source.ExpectStage == ExpectStageIdle {
		// TODO: mark database as idle and drop it later (a month or more later?)
		// we do not remove it immediately because we want to keep the data for a while
		// in case we need to rollback.
		return nil
	}

	dbPassword, err := source.PasswordContent()
	if err != nil {
		return err
	}

	request := &handler.CreateDbRequest{
		InstanceFilter: handler.InstanceFilter{
			Name:   source.InstanceName,
			DbName: source.Name,
		},
		DbOwner:     source.Owner,
		DbPassword:  dbPassword,
		MigrateFrom: source.MigrateFrom,
		Reason:      fmt.Sprintf("Create database %s for source %s", source.Name, source.Type),
	}

	if _, err := h.dbManager.CreateDb(request, false); err != nil {
		h.retryNextTime(source)
		return err
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
		if !oldSource.IsChanged(source) {
			log.Debug().Str("source", source.Name).Msg("Source not changed, skip")
			return nil
		}
		log.Debug().Str("Name", source.Name).Msg("source changed")
	} else {
		log.Debug().Str("Name", source.Name).Msg("source added")
	}

	if source.ExpectStage == "" {
		source.ExpectStage = ExpectStageReady
	}
	h.Databases[source.Name] = source

	go h.sourceProducer.Send(source)

	return nil
}

func (h *SourceHandler) MarkDatabaseSourceIdle(name string) error {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()

	if source, ok := h.Databases[name]; ok {
		if source.ExpectStage != ExpectStageIdle {
			source.ExpectStage = ExpectStageIdle
			source.CronScheduleAt = time.Now().Add(h.Config.DeleyDelete)
			h.cronProducer.Send(&server.CronElement{
				TriggerAt: source.CronScheduleAt,
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
		if source.ExpectStage == ExpectStageIdle && source.CronScheduleAt.Equal(triggerAt) {
			go h.sourceProducer.Send(source)
		}
	}
}

func (h *SourceHandler) OnDbStatusChanged(dbStatus *handler.DbStatusResponse) bool {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()

	if source, ok := h.Databases[dbStatus.Name]; ok {
		if dbStatus.Stage == "DropCompleted" {
			if source.ExpectStage == ExpectStageIdle && source.Synced {
				delete(h.Databases, dbStatus.Name)
			}
		}
		if dbStatus.Stage == source.ExpectStage {
			source.Synced = true
			source.UpdatedAt = dbStatus.UpdatedAt
			source.ResetRetryDelay()
		}
		if dbStatus.IsFailed() {
			h.retryNextTime(source)
		}
	}
	return true
}

// return true if the source is synced else false
func (h *SourceHandler) syncDatabaseSource(source *DatabaseSource) bool {
	dbRequest := &handler.DbRequest{
		InstanceFilter: handler.InstanceFilter{
			Name:   source.InstanceName,
			DbName: source.Name,
		},
	}
	dbStatus, err := h.dbManager.GetDbStatus(dbRequest)
	if err != nil {
		return false
	}

	if dbStatus.Stage == "DropCompleted" {
		if source.ExpectStage == ExpectStageIdle && source.Synced {
			delete(h.Databases, dbStatus.Name)
			return true
		}
	}
	if dbStatus.Stage == source.ExpectStage {
		source.Synced = true
		source.UpdatedAt = dbStatus.UpdatedAt
		return true
	}
	return false
}

func (h *SourceHandler) retryNextTime(source *DatabaseSource) {
	source.CronScheduleAt = time.Now().Add(source.NextRetryDelay())
	log.Debug().Int("RetryDelay", source.RetryDelay).
		Int("RetryTimes", source.RetryTimes).
		Str("DbName", source.Name).
		Msg("Retry next time")

	h.cronProducer.Send(&server.CronElement{
		TriggerAt: source.CronScheduleAt,
		HandleFunc: func(triggerAt time.Time) {
			h.databasesMutex.Lock()
			defer h.databasesMutex.Unlock()
			if source_, ok := h.Databases[source.Name]; ok {
				if source_.CronScheduleAt.Equal(triggerAt) {
					go h.sourceProducer.Send(source_)
				}
			}
		},
	})
}
