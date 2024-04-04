package config

type Config struct {
	Web          WebConfig          `mapstructure:"web" json:"web"`
	Db           DbConfig           `mapstructure:"db" json:"db"`
	Log          LogConfig          `mapstructure:"log" json:"log"`
	RemoteHelper RemoteHelperConfig `mapstructure:"remote_helper" json:"remote_helper"`
}
