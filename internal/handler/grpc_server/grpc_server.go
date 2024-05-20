package grpc_server

import (
	"context"
	"net"
	"time"

	"github.com/a-light-win/pg-helper/api/proto"
	config "github.com/a-light-win/pg-helper/internal/config/server"
	grpcAuth "github.com/a-light-win/pg-helper/pkg/auth/grpc"
	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type GrpcServer struct {
	Config *config.GrpcConfig

	GrpcServer *grpc.Server
	Auth       *grpcAuth.GrpcAuth

	SvcHandler *DbTaskSvcHandler
	QuitCtx    context.Context
}

func NewGrpcServer(config *config.GrpcConfig, quitCtx context.Context) *GrpcServer {
	s := &GrpcServer{Config: config, QuitCtx: quitCtx}

	s.Auth = grpcAuth.NewGrpcAuth(&config.Auth)

	opts := []grpc.ServerOption{}
	creds, err := s.Config.Tls.Credentials()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to get grpc tls credentials")
	}
	opts = append(opts, grpc.Creds(creds))

	opts = append(opts, grpc.UnaryInterceptor(s.Auth.Interceptor))
	opts = append(opts, grpc.StreamInterceptor(s.Auth.StreamInterceptor))

	keepalivePolicy := keepalive.EnforcementPolicy{
		MinTime:             10 * time.Second,
		PermitWithoutStream: false,
	}
	opts = append(opts, grpc.KeepaliveEnforcementPolicy(keepalivePolicy))

	s.SvcHandler = NewDbTaskSvcHandler(config, s.QuitCtx)

	s.GrpcServer = grpc.NewServer(opts...)
	proto.RegisterDbTaskSvcServer(s.GrpcServer, s.SvcHandler)

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

func (s *GrpcServer) Init(setter handler.GlobalSetter) error {
	return nil
}

func (s *GrpcServer) PostInit(getter handler.GlobalGetter) error {
	return nil
}
