package config

type ServerConfig struct {
	Web  WebConfig        `embed:"" prefix:"web-" group:"web"`
	Grpc GrpcServerConfig `embed:"" prefix:"grpc-" group:"grpc"`
}
