package config

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

type GrpcClientConfig struct {
	Tls TlsClientConfig `embed:"" prefix:"tls-" group:"grpc-tls"`

	Url        string `validate:"required,grpcurl" help:"The url that grpc client to connect to"`
	ServerName string `validate:"omitempty,fqdn" help:"The server name that grpc client to connect to"`

	AuthTokenFile string `validate:"omitempty,file" group:"grpc-auth"`
}

type GrpcServerConfig struct {
	Tls TlsServerConfig `embed:"" prefix:"tls-" group:"grpc-tls"`

	BearerAuthEnabled     bool   `default:"false" validate:"required_one_if=true JwtEcDSAVerifyKeyFile JwtEdDSAVerifyKeyFile" group:"grpc-auth"`
	JwtEcDSAVerifyKeyFile string `name:"jwt-ecdsa-verify-key-file" validate:"omitempty,file" group:"grpc-auth"`
	JwtEdDSAVerifyKeyFile string `name:"jwt-eddsa-verify-key-file" validate:"omitempty,file" group:"grpc-auth"`

	Host string `validate:"omitempty,ip" help:"The host that grpc server to listen on"`
	Port int16  `default:"443" help:"The port that grpc server to listen on"`
}

func (g *GrpcClientConfig) AuthToken() (string, error) {
	token, err := os.ReadFile(g.AuthTokenFile)
	if err != nil {
		return "", err
	}
	return string(token), nil
}

func (g *GrpcServerConfig) JwtEcDSAVerifyKey() (interface{}, error) {
	return loadPublicKey(g.JwtEcDSAVerifyKeyFile)
}

func (g *GrpcServerConfig) JwtEdDSAVerifyKey() (interface{}, error) {
	return loadPublicKey(g.JwtEdDSAVerifyKeyFile)
}

func (g *GrpcServerConfig) ListenOn() string {
	return fmt.Sprintf("%s:%d", g.Host, g.Port)
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
