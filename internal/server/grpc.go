package server

import (
	"errors"
	"net"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/server/grpc_handler"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func (s *Server) initGrpc() error {
	opts := []grpc.ServerOption{}

	creds, err := s.Config.Grpc.Tls.Credentials()
	if err != nil {
		log.Error().Err(err).Msg("failed to get grpc tls credentials")
		return err
	}
	opts = append(opts, grpc.Creds(creds))

	hasAuth := false
	if s.Config.Grpc.Tls.MTLSEnabled {
		log.Log().Msg("MTLS is enabled")
		opts = append(opts, grpc.UnaryInterceptor(grpc_handler.TlsAuthInterceptor))
		hasAuth = true
	}

	if !hasAuth && s.Config.Grpc.BearerAuthEnabled {
		log.Log().Msg("Bearer auth is enabled")
		opts = append(opts, grpc.UnaryInterceptor(grpc_handler.BearerAuthInterceptor))
		hasAuth = true
	}

	if !hasAuth {
		err := errors.New("auth method is not provided")
		log.Error().Err(err).Msg("Failed to init grpc server")
		return err
	}

	s.GrpcServer = grpc.NewServer(opts...)
	proto.RegisterDbTaskSvcServer(s.GrpcServer, &grpc_handler.DbTaskSvcHandler{})
	return nil
}

func (s *Server) runGrpcServer() {
	lis, err := net.Listen("tcp", s.Config.Grpc.ListenOn())
	if err != nil {
		log.Fatal().Err(err).
			Str("Host", s.Config.Grpc.Host).
			Int16("Port", s.Config.Grpc.Port).
			Msg("failed to setup grpc server")
		return
	}

	log.Log().Msg("Starting the grpc server")

	if err := s.GrpcServer.Serve(lis); err != nil {
		log.Fatal().Err(err).Msg("failed to run grpc server")
	}
}
