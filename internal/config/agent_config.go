package config

type AgentConfig struct {
	Db   DbConfig         `embed:"" prefix:"db-" group:"db"`
	Grpc GrpcClientConfig `embed:"" prefix:"grpc-" group:"grpc"`
}
