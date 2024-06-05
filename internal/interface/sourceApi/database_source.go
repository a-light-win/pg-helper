package sourceApi

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/a-light-win/pg-helper/internal/interface/grpcServerApi"
	"github.com/rs/zerolog/log"
)

type DatabaseRequest struct {
	Name         string `yaml:"name" json:"name" validate:"required,max=63,id" binding:"required,max=63,id" help:"Name of the database"`
	Owner        string `yaml:"owner" json:"owner" validate:"required,max=63,id" binding:"required,max=63,id" help:"Owner of the database"`
	PasswordFile string `yaml:"password_file" json:"-" validate:"required,file" binding:"-" help:"Path to the password file of the database owner"`
	Password     string `yaml:"-" json:"password" binding:"required,min=8,max=256" help:"Password of the database owner"`

	InstanceName string `yaml:"instance_name" json:"instance_name" validate:"required,max=63,iname" binding:"required,max=63,iname" help:"Name of the pg instance"`
	MigrateFrom  string `yaml:"migrate_from" json:"migrate_from" validate:"omitempty,max=63,iname" binding:"omitempty,max=63,iname" help:"Migrate database from another pg instance"`
	BackupPath   string `yaml:"backup_path" json:"-" validate:"omitempty,file" binding:"-" help:"Path to the backup file"`
}

type DatabaseSource struct {
	*DatabaseRequest

	Type SourceType `yaml:"-"`

	DatabaseSourceStatus
}

type DatabaseSourceStatus struct {
	ExpectState SourceState `yaml:"-"`
	State       SourceState `yaml:"-"`
	UpdatedAt   time.Time   `yaml:"-"`

	RetryDelay int `yaml:"-"`
	RetryTimes int `yaml:"-"`

	LastErrorMsg    string    `yaml:"-"`
	LastScheduledAt time.Time `yaml:"-"`
	NextScheduleAt  time.Time `yaml:"-"`
}

type (
	SourceType  string
	SourceState string
)

const (
	FileSource SourceType = "file"
	WebSource  SourceType = "web"

	SourceStateUnknown    SourceState = "Unknown"
	SourceStatePending    SourceState = "Pending"
	SourceStateScheduling SourceState = "Scheduling"
	SourceStateProcessing SourceState = "Processing"
	SourceStateIdle       SourceState = "Idle"
	SourceStateReady      SourceState = "Ready"
	SourceStateFailed     SourceState = "Failed"
	SourceStateDropped    SourceState = "Dropped"
)

func (s *DatabaseRequest) IsConfigChanged(newSource *DatabaseRequest) bool {
	return s.InstanceName != newSource.InstanceName ||
		s.MigrateFrom != newSource.MigrateFrom ||
		s.BackupPath != newSource.BackupPath
}

func (s *DatabaseRequest) GetName() string {
	return s.Name
}

func (s *DatabaseRequest) PasswordContent() (string, error) {
	if s.Password != "" {
		return s.Password, nil
	}
	if s.PasswordFile != "" {
		password, err := os.ReadFile(s.PasswordFile)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read the password file")
			return "", err
		}
		return strings.TrimSpace(string(password)), nil
	}
	// TODO: allow empty password here,
	// because the password is not required for some cases
	return "", errors.New("password is empty")
}

func (s *DatabaseSource) ResetRetryDelay() {
	log.Debug().Str("DbName", s.Name).Msg("Reset database source retry delay")
	s.RetryTimes = 0
	s.RetryDelay = 0
	s.NextScheduleAt = time.Time{}
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

func (s *DatabaseSource) UpdateState(dbStatus *grpcServerApi.DbStatusResponse) bool {
	if dbStatus.InstanceName != s.InstanceName {
		log.Debug().
			Str("DbName", dbStatus.Name).
			Str("InstanceNameFromStatus", dbStatus.InstanceName).
			Str("InstanceName", s.InstanceName).
			Msg("Ignore the database status from another instance")
		return false
	}

	if dbStatus.Stage == "DropCompleted" && s.ExpectState == SourceStateIdle {
		s.State = SourceStateDropped
		s.UpdatedAt = dbStatus.UpdatedAt

		log.Info().
			Str("DbName", dbStatus.Name).
			Msg("Database dropped")

		return true
	}

	if dbStatus.Stage == string(s.ExpectState) && s.State != s.ExpectState {
		s.State = s.ExpectState
		s.UpdatedAt = dbStatus.UpdatedAt
		s.ResetRetryDelay()
		return true
	}

	if dbStatus.IsFailed() && s.State == SourceStateProcessing {
		s.State = SourceStateFailed
		s.UpdatedAt = dbStatus.UpdatedAt
		s.LastErrorMsg = dbStatus.ErrorMsg
		return true
	}

	return false
}

func (s *DatabaseSource) Synced() bool {
	return s.State == s.ExpectState
}
