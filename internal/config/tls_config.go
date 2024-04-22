package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	ErrEmptyClientCert = errors.New("empty client cert")
	ErrEmptyClientKey  = errors.New("empty client key")
)

var (
	ErrEmptyServerCert           = errors.New("empty server cert")
	ErrEmptyServerKey            = errors.New("empty server key")
	ErrEmptyServerTrustedCaCerts = errors.New("empty server trusted ca certs")
)

type TlsClientConfig struct {
	Enabled              bool   `default:"true" negatable:"true" help:"Enable Tls"`
	ClientTrustedCaCerts string `validate:"omitempty,file" help:"Path to the client trusted ca certs"`

	MTLSEnabled bool   `name:"mtls-enabled" default:"false" negatable:"true" help:"Enable mutual tls" group:"grpc-mtls"`
	ClientCert  string `validate:"required_if=MTLSEnabled true,omitempty,file" help:"Path to the client tls cert" group:"grpc-mtls"`
	ClientKey   string `validate:"required_if=MTLSEnabled true,omitempty,file" help:"Path to the client tls key" group:"grpc-mtls"`
}

type TlsServerConfig struct {
	Enabled    bool   `default:"true" negatable:"true" help:"Enable Tls"`
	ServerCert string `validate:"required_if=Enabled true,omitempty,file" help:"Path to the server tls cert"`
	ServerKey  string `validate:"required_if=Enabled true,omitempty,file" help:"Path to the server tls key"`

	MTLSEnabled          bool   `name:"mtls-enabled" negatable:"true" help:"Enable mutual tls" group:"grpc-mtls"`
	ServerTrustedCaCerts string `validate:"required_if=MTLSEnabled true,omitempty,file" help:"Path to the server trusted ca certs" group:"grpc-mtls"`
	TrustedClientDomain  string `validate:"omitempty,fqdn" group:"grpc-mtls"`
}

func (t *TlsClientConfig) TlsConfig() (*tls.Config, error) {
	if !t.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	if t.MTLSEnabled {
		cert, err := loadCert(t.ClientCert, t.ClientKey)
		if err != nil {
			return nil, err
		}
		if cert != nil {
			tlsConfig.Certificates = []tls.Certificate{*cert}
		}
	}

	if t.ClientTrustedCaCerts != "" {
		ca, err := loadCA(t.ClientTrustedCaCerts)
		if err != nil {
			return nil, err
		}
		tlsConfig.RootCAs = ca
	}

	return tlsConfig, nil
}

func (t *TlsClientConfig) Credentials() (credentials.TransportCredentials, error) {
	if !t.Enabled {
		return insecure.NewCredentials(), nil
	}

	tlsConfig, err := t.TlsConfig()
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(tlsConfig), nil
}

func (t *TlsServerConfig) Credentials() (credentials.TransportCredentials, error) {
	if !t.Enabled {
		return insecure.NewCredentials(), nil
	}

	tlsConfig, err := t.TlsConfig()
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(tlsConfig), nil
}

func (t *TlsServerConfig) TlsConfig() (*tls.Config, error) {
	if !t.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{}
	cert, err := loadCert(t.ServerCert, t.ServerKey)
	if err != nil {
		return nil, err
	}
	if cert != nil {
		tlsConfig.Certificates = []tls.Certificate{*cert}
	}

	if t.ServerTrustedCaCerts != "" {
		ca, err := loadCA(t.ServerTrustedCaCerts)
		if err != nil {
			return nil, err
		}
		tlsConfig.ClientCAs = ca
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return tlsConfig, nil
}

func loadCert(certFile string, keyFile string) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Error().Err(err).Str("CertFile", certFile).Str("KeyFile", keyFile).Msg("Failed to load certificate")
		return nil, err
	}
	return &cert, nil
}

func loadCA(caCertFile string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load CA certificate")
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool, nil
}
