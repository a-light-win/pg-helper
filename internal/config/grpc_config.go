package config

import (
	"os"
)

type GrpcClientConfig struct {
	Tls TlsClientConfig `embed:"" prefix:"tls-" group:"grpc-tls"`

	Url        string `validate:"required,grpcurl" help:"The url that grpc client to connect to"`
	ServerName string `validate:"omitempty,fqdn" help:"The server name that grpc client to connect to"`

	AuthTokenFile string `validate:"omitempty,file" env:"PG_HELPER_GRPC_AUTH_TOKEN_FILE" group:"grpc-auth"`
}

func (g *GrpcClientConfig) AuthToken() (string, error) {
	token, err := os.ReadFile(g.AuthTokenFile)
	if err != nil {
		return "", err
	}
	return string(token), nil
}
