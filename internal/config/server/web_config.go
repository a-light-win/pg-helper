package server

type WebConfig struct {
	Enabled        bool     `default:"false" help:"Enable the web server"`
	UseH2C         bool     `name:"use-h2c" default:"true" negatable:"true" help:"Enable http/2 without TLS"`
	TrustedProxies []string `help:"Addresses of the trusted proxies"`

	Tls TlsConfig `embed:"" prefix:"tls-" group:"web-tls"`
}

func (c *WebConfig) ListenOn() string {
	// TODO: customize the address and port
	return ":8080"
}
