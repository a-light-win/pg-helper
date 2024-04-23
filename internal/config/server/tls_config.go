package server

import (
	"crypto/tls"

	"github.com/a-light-win/pg-helper/internal/utils"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type TlsConfig struct {
	Enabled    bool   `default:"true" negatable:"true" help:"Enable Tls"`
	ServerCert string `validate:"required_if=Enabled true,omitempty,file" help:"Path to the server tls cert"`
	ServerKey  string `validate:"required_if=Enabled true,omitempty,file" help:"Path to the server tls key"`

	MTLSEnabled          bool   `name:"mtls-enabled" negatable:"true" help:"Enable mutual tls" group:"grpc-mtls"`
	ServerTrustedCaCerts string `validate:"required_if=MTLSEnabled true,omitempty,file" help:"Path to the server trusted ca certs" group:"grpc-mtls"`
	TrustedClientDomain  string `validate:"omitempty,fqdn" group:"grpc-mtls"`
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

func (t *TlsConfig) TlsConfig() (*tls.Config, error) {
	if !t.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{}
	cert, err := utils.LoadCert(t.ServerCert, t.ServerKey)
	if err != nil {
		return nil, err
	}
	if cert != nil {
		tlsConfig.Certificates = []tls.Certificate{*cert}
	}

	if t.ServerTrustedCaCerts != "" {
		ca, err := utils.LoadCA(t.ServerTrustedCaCerts)
		if err != nil {
			return nil, err
		}
		tlsConfig.ClientCAs = ca
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return tlsConfig, nil
}
