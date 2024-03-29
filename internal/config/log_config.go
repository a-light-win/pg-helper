package config

import "github.com/rs/zerolog"

type LogConfig struct {
	Level string `mapstructure:"level" json:"level"`
}

func (c *LogConfig) GetLevel() zerolog.Level {
	l, err := zerolog.ParseLevel(c.Level)
	if err != nil {
		return zerolog.WarnLevel
	}
	return l
}
