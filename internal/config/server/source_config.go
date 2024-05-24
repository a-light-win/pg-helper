package server

import "time"

type SourceConfig struct {
	File FileSourceConfig `embed:"" prefix:"file-" group:"file-source"`

	DeleyDelete time.Duration `default:"5s"`
}
