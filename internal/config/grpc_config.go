package config

type GrpcClientConfig struct {
	TlsClientConfig

	Url        string
	ServerName string
}

type GrpcServerConfig struct{}
