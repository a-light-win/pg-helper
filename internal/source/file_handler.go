package source

import (
	"errors"
	"os"
	"strings"

	config "github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/a-light-win/pg-helper/pkg/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
)

const FileSourceEnding = ".yaml"

type FileSourceHandler struct {
	ParentSourceHandler

	Config *config.FileSourceConfig

	sourceMap map[string]string
}

func NewFileSourceHandler(handler *SourceHandler) *FileSourceHandler {
	return &FileSourceHandler{
		ParentSourceHandler: handler,
		Config:              &handler.Config.File,
		sourceMap:           make(map[string]string),
	}
}

func (h *FileSourceHandler) Init(setter server.GlobalSetter) error {
	if !h.Config.Enabled {
		log.Log().Msg("File source handler is disabled")
		return nil
	}

	if len(h.Config.FilePaths) == 0 {
		return errors.New("no file paths provided")
	}

	return nil
}

func (h *FileSourceHandler) PostInit(getter server.GlobalGetter) error {
	if h.Config.Enabled {
		for _, path := range h.Config.FilePaths {
			if err := h.loadDatabaseSources(path); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *FileSourceHandler) Handle(msg server.NamedElement) error {
	event := msg.(*server.NamedFileEvent)
	switch event.Op {
	case fsnotify.Create, fsnotify.Write:
		h.loadDatabaseSourceFromFile(event.Name)
	case fsnotify.Remove, fsnotify.Rename:
		h.removeDatabaseSource(event.Name)
	}
	return nil
}

func (h *FileSourceHandler) FilesToWatch() []string {
	if !h.Config.Enabled {
		return nil
	}
	return h.Config.FilePaths
}

func (h *FileSourceHandler) OnWatchError(err error) {
	log.Error().Err(err).Msg("File monitor error")
}

func (h *FileSourceHandler) loadDatabaseSources(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Warn().Err(err).Str("path", path).Msg("Load database source failed")
		return err
	}

	if fileInfo.IsDir() {
		return h.loadDatabaseSourcesFromDir(path)
	} else {
		return h.loadDatabaseSourceFromFile(path)
	}
}

func (h *FileSourceHandler) loadDatabaseSourcesFromDir(path string) error {
	dir, err := os.ReadDir(path)
	if err != nil {
		log.Warn().Err(err).Str("path", path).Msg("Load database source from dir failed")
		return err
	}

	for _, file := range dir {
		if file.IsDir() {
			// We don't support sub directory for simplicity
			log.Debug().Str("path", path+"/"+file.Name()).
				Msg("Load database sources from sub directory is not supported")
			continue
		} else {
			if err := h.loadDatabaseSourceFromFile(path + "/" + file.Name()); err != nil {
				continue
			}
		}
	}
	return nil
}

func (h *FileSourceHandler) loadDatabaseSourceFromFile(path string) error {
	if !strings.HasSuffix(path, FileSourceEnding) {
		log.Debug().Str("path", path).Msg("Skip file not end with .yaml")
		return nil
	}

	var fileSource DatabaseSource
	if err := utils.LoadYaml(path, &fileSource); err != nil {
		log.Warn().Err(err).Str("path", path).Msg("Load database source from file failed")
		return err
	}

	h.sourceMap[path] = fileSource.Name
	fileSource.Type = FileSource
	fileSource.Synced = false

	if err := h.AddDatabaseSource(&fileSource); err != nil {
		log.Warn().Err(err).Str("path", path).Msg("Add database source failed")
		return err
	}

	return nil
}

func (h *FileSourceHandler) removeDatabaseSource(path string) error {
	name, ok := h.sourceMap[path]
	if !ok {
		log.Debug().Str("path", path).Msg("Skip file not in source map")
		return nil
	}

	delete(h.sourceMap, path)

	if err := h.MarkDatabaseSourceIdle(name); err != nil {
		return err
	}

	return nil
}
