package server

import (
	"fmt"

	"github.com/a-light-win/pg-helper/pkg/auth"
)

type GrpcConfig struct {
	Tls  TlsConfig       `embed:"" prefix:"tls-" group:"grpc-tls"`
	Auth auth.AuthConfig `embed:"" prefix:"auth-" group:"grpc-auth"`

	Host string `validate:"omitempty,ip" help:"The host that grpc server to listen on"`
	Port int16  `default:"443" help:"The port that grpc server to listen on"`
}

func (g *GrpcConfig) ListenOn() string {
	return fmt.Sprintf("%s:%d", g.Host, g.Port)
}
