package server

import (
	"fmt"

	"github.com/a-light-win/pg-helper/internal/utils"
)

type GrpcConfig struct {
	Tls TlsConfig `embed:"" prefix:"tls-" group:"grpc-tls"`

	BearerAuthEnabled     bool   `default:"false" validate:"required_one_if=true JwtEcDSAVerifyKeyFile JwtEdDSAVerifyKeyFile" group:"grpc-auth"`
	JwtEcDSAVerifyKeyFile string `name:"jwt-ecdsa-verify-key-file" validate:"omitempty,file" group:"grpc-auth"`
	JwtEdDSAVerifyKeyFile string `name:"jwt-eddsa-verify-key-file" validate:"omitempty,file" group:"grpc-auth"`

	Host string `validate:"omitempty,ip" help:"The host that grpc server to listen on"`
	Port int16  `default:"443" help:"The port that grpc server to listen on"`
}

func (g *GrpcConfig) JwtEcDSAVerifyKey() (interface{}, error) {
	return utils.LoadPublicKey(g.JwtEcDSAVerifyKeyFile)
}

func (g *GrpcConfig) JwtEdDSAVerifyKey() (interface{}, error) {
	return utils.LoadPublicKey(g.JwtEdDSAVerifyKeyFile)
}

func (g *GrpcConfig) ListenOn() string {
	return fmt.Sprintf("%s:%d", g.Host, g.Port)
}
