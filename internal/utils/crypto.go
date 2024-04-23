package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

func LoadPublicKey(file string) (interface{}, error) {
	// Read the file
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// Decode the PEM block
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	// Parse the public key
	return x509.ParsePKIXPublicKey(block.Bytes)
}

func LoadCert(certFile string, keyFile string) (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Error().Err(err).Str("CertFile", certFile).Str("KeyFile", keyFile).Msg("Failed to load certificate")
		return nil, err
	}
	return &cert, nil
}

func LoadCA(caCertFile string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load CA certificate")
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool, nil
}
