package config

type Config struct {
	Web WebConfig `mapstructure:"web" json:"web"`
	Db  DbConfig  `mapstructure:"db" json:"db"`
}
