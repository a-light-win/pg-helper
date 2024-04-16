package config

import (
	"crypto/tls"
	"crypto/x509"
	"os"

	"github.com/rs/zerolog/log"
)

type TlsClientConfig struct {
	Enabled              bool   `default:"true" `
	ClientCert           string `default:"/etc/pg-helper/certs/client.crt"`
	ClientKey            string `default:"/etc/pg-helper/certs/client.key"`
	ClientTrustedCaCerts string `default:"/etc/pg-helper/certs/client-trusted-ca.crt"`
}

type TlsServerConfig struct {
	Enabled              bool   `default:"true"`
	ServerCert           string `default:"/etc/pg-helper/certs/server.crt"`
	ServerKey            string `default:"/etc/pg-helper/certs/server.key"`
	ServerTrustedCaCerts string `default:"/etc/pg-helper/certs/server-trusted-ca.crt"`
}

func (t *TlsClientConfig) TlsConfig() (*tls.Config, error) {
	if !t.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	cert, err := loadCert(t.ClientCert, t.ClientKey)
	if err != nil {
		return nil, err
	}
	if cert != nil {
		tlsConfig.Certificates = []tls.Certificate{*cert}
	}

	ca, err := loadCA(t.ClientTrustedCaCerts)
	if err != nil {
		return nil, err
	}
	if ca != nil {
		tlsConfig.RootCAs = ca
	}

	return tlsConfig, nil
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
	ca, err := loadCA(t.ServerTrustedCaCerts)
	if err != nil {
		return nil, err
	}
	if ca != nil {
		tlsConfig.ClientCAs = ca
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	return tlsConfig, nil
}

func loadCert(certFile string, keyFile string) (*tls.Certificate, error) {
	if certFile == "" || keyFile == "" {
		return nil, nil
	}

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		log.Error().Err(err).Msg("Failed to load client certificate")
		return nil, err
	}

	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		log.Error().Err(err).Msg("Failed to load client certificate")
		return nil, err
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load client certificate")
		return nil, err
	}
	return &cert, nil
}

func loadCA(caCertFile string) (*x509.CertPool, error) {
	if caCertFile == "" {
		return nil, nil
	}
	if _, err := os.Stat(caCertFile); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		log.Error().Err(err).Msg("Failed to load CA certificate")
		return nil, err
	}

	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load CA certificate")
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool, nil
}
