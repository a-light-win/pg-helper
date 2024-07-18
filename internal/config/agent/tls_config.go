package agent

import (
	"crypto/tls"

	"github.com/a-light-win/pg-helper/pkg/utils"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type TlsConfig struct {
	Enabled              bool   `default:"false" help:"Enable Tls"`
	ClientTrustedCaCerts string `validate:"omitempty,file" help:"Path to the client trusted ca certs"`

	MTLSEnabled bool   `name:"mtls-enabled" default:"false" help:"Enable mutual tls" group:"grpc-mtls"`
	ClientCert  string `validate:"required_if=MTLSEnabled true,omitempty,file" help:"Path to the client tls cert" group:"grpc-mtls"`
	ClientKey   string `validate:"required_if=MTLSEnabled true,omitempty,file" help:"Path to the client tls key" group:"grpc-mtls"`
}

func (t *TlsConfig) TlsConfig() (*tls.Config, error) {
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

func (t *TlsConfig) Credentials() (credentials.TransportCredentials, error) {
	if !t.Enabled {
		return insecure.NewCredentials(), nil
	}

	tlsConfig, err := t.TlsConfig()
	if err != nil {
		return nil, err
	}
	return credentials.NewTLS(tlsConfig), nil
}
