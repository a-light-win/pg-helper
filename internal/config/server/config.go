package server

type ServerConfig struct {
	Web  WebConfig  `embed:"" prefix:"web-" group:"web"`
	Grpc GrpcConfig `embed:"" prefix:"grpc-" group:"grpc"`
	Db   DbConfig   `embed:"" prefix:"db-" group:"db"`
}
