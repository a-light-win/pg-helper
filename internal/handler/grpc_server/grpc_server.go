package grpc_server

import (
	"context"
	"net"

	"github.com/a-light-win/pg-helper/api/proto"
	config "github.com/a-light-win/pg-helper/internal/config/server"
	grpcAuth "github.com/a-light-win/pg-helper/pkg/auth/grpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	Config     *config.GrpcConfig
	GrpcServer *grpc.Server
	Auth       *grpcAuth.GrpcAuth
}

func NewGrpcServer(config *config.GrpcConfig) *GrpcServer {
	s := &GrpcServer{Config: config}

	s.Auth = grpcAuth.NewGrpcAuth(&config.Auth)

	opts := []grpc.ServerOption{}
	creds, err := s.Config.Tls.Credentials()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get grpc tls credentials")
	}
	opts = append(opts, grpc.Creds(creds))

	opts = append(opts, grpc.UnaryInterceptor(s.Auth.Interceptor))
	opts = append(opts, grpc.StreamInterceptor(s.Auth.StreamInterceptor))

	s.GrpcServer = grpc.NewServer(opts...)
	proto.RegisterDbTaskSvcServer(s.GrpcServer, &DbTaskSvcHandler{})

	return s
}

func (s *GrpcServer) Run() {
	lis, err := net.Listen("tcp", s.Config.ListenOn())
	if err != nil {
		log.Fatal().Err(err).
			Str("Host", s.Config.Host).
			Int16("Port", s.Config.Port).
			Msg("failed to setup grpc server")
		return
	}

	log.Log().Msg("Starting the grpc server")

	if err := s.GrpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to run grpc server")
	}
}

func (s *GrpcServer) Shutdown(ctx context.Context) error {
	s.GrpcServer.GracefulStop()
	return nil
}
