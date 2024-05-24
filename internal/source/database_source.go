package source

import "time"

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
	Synced            bool      `yaml:"-"`
	LastScheduledTime time.Time `yaml:"-"`

	ShouldRemove   bool      `yaml:"-"`
	ShouldRemoveAt time.Time `yaml:"-"`
}

type SourceType string

const (
	FileSource SourceType = "file"
)

func (s *DatabaseSource) IsChanged(newSource *DatabaseSource) bool {
	return s.InstanceName != newSource.InstanceName ||
		s.MigrateFrom != newSource.MigrateFrom ||
		s.BackupPath != newSource.BackupPath
}

func (s *DatabaseSource) GetName() string {
	return s.Name
}
