package source

import (
	"sync"
	"time"

	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/internal/constants"
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

	validator *validator.Validate
}

type ParentSourceHandler interface {
	AddDatabaseSource(source *DatabaseSource) error
	MarkDatabaseSourceShouldRemove(name string) error
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
	return nil
}

func (h *SourceHandler) Handle(msg server.NamedElement) error {
	// TODO: handle message
	// source := msg.(*DatabaseSource)
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
	}

	h.Databases[source.Name] = source

	go h.sourceProducer.Send(source)

	return nil
}

func (h *SourceHandler) MarkDatabaseSourceShouldRemove(name string) error {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()

	if source, ok := h.Databases[name]; ok {
		if !source.ShouldRemove {
			source.ShouldRemove = true
			source.ShouldRemoveAt = time.Now().Add(h.Config.DeleyDelete)
			h.cronProducer.Send(&server.CronElement{
				TriggerAt: source.ShouldRemoveAt,
				HandleFunc: func(triggerAt time.Time) {
					h.removeDatabaseSource(name, triggerAt)
				},
			})
		}
	}
	return nil
}

func (h *SourceHandler) removeDatabaseSource(name string, triggerAt time.Time) {
	h.databasesMutex.Lock()
	defer h.databasesMutex.Unlock()
	if source, ok := h.Databases[name]; ok {
		if source.ShouldRemove && source.ShouldRemoveAt.Equal(triggerAt) {
			delete(h.Databases, name)
			go h.sourceProducer.Send(source)
		}
	}
}
