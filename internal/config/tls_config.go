package config

import (
	"crypto/tls"
	"errors"

	"github.com/a-light-win/pg-helper/internal/utils"
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

func (t *TlsClientConfig) TlsConfig() (*tls.Config, error) {
	if !t.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	if t.MTLSEnabled {
		cert, err := utils.LoadCert(t.ClientCert, t.ClientKey)
		if err != nil {
			return nil, err
		}
		if cert != nil {
			tlsConfig.Certificates = []tls.Certificate{*cert}
		}
	}

	if t.ClientTrustedCaCerts != "" {
		ca, err := utils.LoadCA(t.ClientTrustedCaCerts)
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
