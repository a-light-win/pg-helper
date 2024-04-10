package config

type AgentConfig struct {
	Id   string           `help:"Agent ID"`
	Db   DbConfig         `embed:"" prefix:"db-" group:"db"`
	Grpc GrpcClientConfig `embed:"" prefix:"grpc-" group:"grpc"`
}
