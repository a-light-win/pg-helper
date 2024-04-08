package config

type Config struct {
	Web          WebConfig          `mapstructure:"web" json:"web" embed:"true" prefix:"web." group:"web"`
	Db           DbConfig           `mapstructure:"db" json:"db" embed:"true" prefix:"db." group:"db"`
	RemoteHelper RemoteHelperConfig `mapstructure:"remote_helper" json:"remote_helper" embed:"true" prefix:"remote-helper." group:"remote-helper"`
}
