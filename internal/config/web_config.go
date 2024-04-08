package config

type WebConfig struct {
	UseH2C         bool     `json:"use_h2c" name:"use-h2c" default:"true" negatable:"true" help:"Enable http/2 without TLS"`
	TrustedProxies []string `json:"trusted_proxies" help:"Addresses of the trusted proxies"`
}
