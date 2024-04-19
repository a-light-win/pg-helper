package config

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type GrpcClientConfig struct {
	TlsClientConfig

	Url        string
	ServerName string

	AuthTokenFile string `name:"auth-token-file" default:"/etc/pg-helper/auth-token"`
}

type GrpcServerConfig struct {
	TlsServerConfig

	TrustedClientDomain   string
	JwtEcDSAVerifyKeyFile string `name:"jwt-ecdsa-verify-key-file" default:"/etc/pg-helper/certs/jwt-ecdsa-verify-key.pem"`
	JwtEdDSAVerifyKeyFile string `name:"jwt-eddsa-verify-key-file" default:"/etc/pg-helper/certs/jwt-eddsa-verify-key.pem"`
}

func (g *GrpcServerConfig) JwtEcDSAVerifyKey() (interface{}, error) {
	return loadPublicKey(g.JwtEcDSAVerifyKeyFile)
}

func (g *GrpcServerConfig) JwtEdDSAVerifyKey() (interface{}, error) {
	return loadPublicKey(g.JwtEdDSAVerifyKeyFile)
}

func (g *GrpcClientConfig) AuthToken() (string, error) {
	token, err := os.ReadFile(g.AuthTokenFile)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func loadPublicKey(file string) (interface{}, error) {
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
