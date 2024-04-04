package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

// Communication with other pg-helper
type RemoteHelperConfig struct {
	UrlTemplate string `mapstructure:"url_template" json:"url_template"`

	// TLS settings
	// If server enabled mtls, the ClientCertPath and ClientKeyPath must be set
	// so that the client can establish a connection with the server
	ClientCertPath string `mapstructure:"client_cert_path" json:"client_cert_path"`
	ClientKeyPath  string `mapstructure:"client_key_path" json:"client_key_path"`
	// The path of the trusted CA certificate, if not set, the system CA will be used
	// client will verify the server's certificate with the trusted CA certificate
	ClientTrustedCAPath string `mapstructure:"client_trusted_ca_path" json:"client_trusted_ca_path"`

	BearerTokenPath string `mapsctructure:"bearer_token_path" json:"bearer_token_path"`
}

func (c *RemoteHelperConfig) Url(pgVersion int) string {
	return fmt.Sprintf(c.UrlTemplate, pgVersion)
}

func (c *RemoteHelperConfig) TLSConfig() *tls.Config {
	tlsConfig := &tls.Config{}
	if c.ClientCertPath != "" && c.ClientKeyPath != "" {
		cert, err := tls.LoadX509KeyPair(c.ClientCertPath, c.ClientKeyPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to load the client certificate")
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	if c.ClientTrustedCAPath != "" {
		caCert, err := os.ReadFile(c.ClientTrustedCAPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read the trusted CA certificate")
		}
		tlsConfig.RootCAs = x509.NewCertPool()
		tlsConfig.RootCAs.AppendCertsFromPEM(caCert)
	}
	return tlsConfig
}

func (c *RemoteHelperConfig) BearerToken() string {
	if c.BearerTokenPath != "" {
		token, err := os.ReadFile(c.BearerTokenPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to read the bearer token")
		}
		return string(token)
	}
	return ""
}
