package handler

import (
	"context"
	"errors"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

type NamedFileEvent struct{ fsnotify.Event }

func (e *NamedFileEvent) GetName() string {
	return e.Name
}

type FileChangedHandler interface {
	Handler

	FilesToWatch() []string
	OnWatchError(error)
}

type FileMonitor struct {
	Name    string
	Handler FileChangedHandler

	watcher *fsnotify.Watcher
	exited  chan struct{}
}

func (m *FileMonitor) Init(setter GlobalSetter) error {
	if err := m.Handler.Init(setter); err != nil {
		return err
	}

	var err error
	m.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Error().Err(err).Msg("Failed to create watcher")
		return err
	}
	m.exited = make(chan struct{})

	return nil
}

func (m *FileMonitor) PostInit(getter GlobalGetter) error {
	if err := m.Handler.PostInit(getter); err != nil {
		return err
	}

	watchList := m.Handler.FilesToWatch()
	if len(watchList) == 0 {
		err := errors.New("no files to watch")
		return err
	}
	for _, file := range watchList {
		if err := m.watcher.Add(file); err != nil {
			return err
		}
	}

	return nil
}

func (m *FileMonitor) Run() {
	log.Log().Msgf("%s is running", m.Name)

	defer func() {
		m.exited <- struct{}{}
	}()

	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			m.Handler.Handle(&NamedFileEvent{event})
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			m.Handler.OnWatchError(err)
		}
	}
}

func (m *FileMonitor) Shutdown(ctx context.Context) {
	log.Log().Msgf("%s is shuting down", m.Name)

	m.watcher.Close()
	<-m.exited

	if shutdowner, ok := m.Handler.(Shutdowner); ok {
		shutdowner.Shutdown(ctx)
	}

	log.Log().Msgf("%s is shutdown", m.Name)
}
