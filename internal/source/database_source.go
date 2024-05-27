package source

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type DatabaseSource struct {
	Name         string `yaml:"name" json:"name" validate:"required,max=63,id" help:"Name of the database"`
	Owner        string `yaml:"owner" json:"owner" validate:"required,max=63,id" help:"Owner of the database"`
	PasswordFile string `yaml:"password_file" json:"password_file" validate:"required,file" help:"Path to the password file of the database owner"`

	InstanceName string `yaml:"instance_name" json:"instance_name" validate:"required,max=63,iname" help:"Name of the pg instance"`
	MigrateFrom  string `yaml:"migrate_from" json:"migrate_from" validate:"omitempty,max=63,iname" help:"Migrate database from another pg instance"`
	BackupPath   string `yaml:"backup_path" json:"backup_path" validate:"omitempty,file" help:"Path to the backup file"`

	Type SourceType `yaml:"-"`

	DatabaseSourceStatus
}

type DatabaseSourceStatus struct {
	ExpectStage string    `yaml:"-"`
	Synced      bool      `yaml:"-"`
	UpdatedAt   time.Time `yaml:"-"`

	RetryDelay int `yaml:"-"`
	RetryTimes int `yaml:"-"`

	LastScheduledTime time.Time `yaml:"-"`
	CronScheduleAt    time.Time `yaml:"-"`
}

type SourceType string

const (
	FileSource SourceType = "file"

	ExpectStageIdle  string = "Idle"
	ExpectStageReady string = "Ready"
)

func (s *DatabaseSource) IsChanged(newSource *DatabaseSource) bool {
	return s.InstanceName != newSource.InstanceName ||
		s.MigrateFrom != newSource.MigrateFrom ||
		s.BackupPath != newSource.BackupPath
}

func (s *DatabaseSource) GetName() string {
	return s.Name
}

func (s *DatabaseSource) PasswordContent() (string, error) {
	if s.PasswordFile != "" {
		password, err := os.ReadFile(s.PasswordFile)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read the password file")
			return "", err
		}
		return strings.TrimSpace(string(password)), nil
	}
	return "", errors.New("password file is empty")
}

func (s *DatabaseSource) ResetRetryDelay() {
	log.Debug().Str("DbName", s.Name).Msg("Reset database source retry delay")
	s.RetryTimes = 0
	s.RetryDelay = 0
}

func (s *DatabaseSource) NextRetryDelay() time.Duration {
	s.RetryTimes++
	if s.RetryDelay == 0 {
		s.RetryDelay = 2
	} else if s.RetryDelay == 2 {
		s.RetryDelay = 3
	} else if s.RetryDelay == 3 {
		s.RetryDelay = 5
	} else if s.RetryDelay < 3600 {
		s.RetryDelay += s.RetryDelay
	}
	return time.Duration(s.RetryDelay) * time.Second
}
