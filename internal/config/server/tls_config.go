package server

import (
	"crypto/tls"

	"github.com/a-light-win/pg-helper/pkg/utils"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type TlsConfig struct {
	Enabled    bool   `default:"false" help:"Enable Tls"`
	ServerCert string `validate:"required_if=Enabled true,omitempty,file" help:"Path to the server tls cert"`
	ServerKey  string `validate:"required_if=Enabled true,omitempty,file" help:"Path to the server tls key"`

	MTLSEnabled         bool   `name:"mtls-enabled" default:"false" help:"Enable mutual tls"`
	TrustedCA           string `name:"trusted-ca" validate:"required_if=MTLSEnabled true,omitempty,file" help:"Path to the server trusted ca certs"`
	TrustedClientDomain string `validate:"omitempty,fqdn"`
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

	if t.TrustedCA != "" {
		ca, err := utils.LoadCA(t.TrustedCA)
		if err != nil {
			return nil, err
		}
		tlsConfig.ClientCAs = ca
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return tlsConfig, nil
}
