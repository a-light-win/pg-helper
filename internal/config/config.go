package config

type Config struct {
	Web          WebConfig          `mapstructure:"web" json:"web" embed:"" prefix:"web-" group:"web"`
	Db           DbConfig           `mapstructure:"db" json:"db" embed:"" prefix:"db-" group:"db"`
	RemoteHelper RemoteHelperConfig `mapstructure:"remote_helper" json:"remote_helper" embed:"" prefix:"remote-helper-" group:"remote-helper"`
}
