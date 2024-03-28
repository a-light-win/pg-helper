package config

type WebConfig struct {
	UseH2C         bool     `mapstructure:"use_h2c" json:"use_h2c"`
	TrustedProxies []string `mapstructure:"trusted_proxies" json:"trusted_proxies"`
}
