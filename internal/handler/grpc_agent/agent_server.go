package grpc_agent

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"time"

	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/a-light-win/pg-helper/pkg/proto"
	"github.com/a-light-win/pg-helper/pkg/server"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

type GrpcAgentServer struct {
	GrpcConfig *config.GrpcClientConfig
	GrpcClient proto.DbJobSvcClient

	QuitCtx context.Context

	DbApi *db.DbApi

	handler     *GrpcAgentHandler
	jobProducer server.Producer

	exited chan struct{}
}

func NewGrpcAgentServer(grpcConfig *config.GrpcClientConfig, quitCtx context.Context) *GrpcAgentServer {
	server := &GrpcAgentServer{
		GrpcConfig: grpcConfig,
		QuitCtx:    quitCtx,
		exited:     make(chan struct{}),
	}
	return server
}

func (s *GrpcAgentServer) initGrpcClient() error {
	dialOptions := []grpc.DialOption{}

	creds, err := s.GrpcConfig.Tls.Credentials()
	if err != nil {
		return err
	}
	dialOptions = append(dialOptions, grpc.WithTransportCredentials(creds))

	if s.GrpcConfig.ServerName != "" {
		dialOptions = append(dialOptions, grpc.WithAuthority(s.GrpcConfig.ServerName))
	}

	authToken, err := s.GrpcConfig.AuthToken()
	if err == nil {
		log.Log().Msg("Auth token is provided")
		authCreds := NewAuthToken(authToken, s.GrpcConfig.Tls.Enabled)
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(authCreds))
	} else if !s.GrpcConfig.Tls.MTLSEnabled {
		err := errors.New("auth method is not provided")
		log.Error().Err(err).Msg("Failed to init grpc client")
		return err
	}

	ka := keepalive.ClientParameters{
		Time:                time.Second * 15,
		Timeout:             time.Second * 5,
		PermitWithoutStream: false,
	}
	dialOptions = append(dialOptions, grpc.WithKeepaliveParams(ka))

	conn, err := grpc.Dial(s.GrpcConfig.Url, dialOptions...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to dial gRPC server")
		return err
	}

	s.GrpcClient = proto.NewDbJobSvcClient(conn)
	return nil
}

func (s *GrpcAgentServer) Init(setter server.GlobalSetter) error {
	if err := s.initGrpcClient(); err != nil {
		return err
	}
	setter.Set(constants.AgentKeyGrpcClient, s.GrpcClient)

	return nil
}

func (s *GrpcAgentServer) PostInit(getter server.GlobalGetter) error {
	s.DbApi = getter.Get(constants.AgentKeyDbApi).(*db.DbApi)
	s.jobProducer = getter.Get(constants.AgentKeyJobProducer).(server.Producer)

	s.handler = NewGrpcAgentHandler(s.DbApi, s.GrpcClient, s.jobProducer, s.QuitCtx)

	return nil
}

func (s *GrpcAgentServer) Run() {
	defer func() {
		s.exited <- struct{}{}
	}()

	service := s.registerService()
	if service == nil {
		return
	}

	for {

		select {
		case <-s.QuitCtx.Done():
			return
		default:
		}

		task, err := service.Recv()
		if err != nil {
			status, ok := status.FromError(err)
			if ok && status.Code() == codes.Canceled {
				return
			}
			if ok && status.Code() == codes.Unauthenticated {
				log.Error().Err(err).Msg("Failed to receive task")
				return
			}

			if err == io.EOF {
				log.Info().Msg("Grpc server closed the connection")
			} else {
				log.Warn().Err(err).Msg("Failed to receive task")
			}

			wait := time.Duration(rand.Intn(3)+1) * time.Second
			time.Sleep(wait)

			service = s.registerService()
			if service == nil {
				return
			}
			continue
		}

		go s.handler.handle(task)
	}
}

func (s *GrpcAgentServer) Shutdown(ctx context.Context) {
	log.Log().Msg("Shutting down gRPC agent server")

	<-s.exited

	log.Log().Msg("gRPC agent server is down")
}

func (s *GrpcAgentServer) registerService() proto.DbJobSvc_RegisterClient {
	wait := 2

	registerAgentLoader := &registerAgentLoader{agent: s}
	if !s.runUntilSuccess(registerAgentLoader, wait) {
		return nil
	}

	grpcServiceLoader := &grpcServiceLoader{agent: s, registerAgent: registerAgentLoader.registerAgent}
	if !s.runUntilSuccess(grpcServiceLoader, wait) {
		return nil
	}
	return grpcServiceLoader.service
}

func (s *GrpcAgentServer) runUntilSuccess(runer utils.Runner, firstWait int) bool {
	continueWait := firstWait
	maxContinueWait := 60

	for {

		select {
		case <-s.QuitCtx.Done():
			return false
		default:
			if runer.Run() {
				return true
			}
		}

		wait := time.Duration(continueWait) * time.Second
		if continueWait < maxContinueWait {
			continueWait += 1
		}

		select {
		case <-s.QuitCtx.Done():
			return false
		case <-time.After(wait):
			continue
		}
	}
}

func (s *GrpcAgentServer) loadRegisterAgent() (*proto.RegisterInstance, error) {
	registerAgent := &proto.RegisterInstance{
		Name:      s.DbApi.DbConfig.InstanceName,
		PgVersion: s.DbApi.DbConfig.CurrentVersion,
	}

	if dbs, err := s.DbApi.ListDbs(nil); err != nil {
		log.Error().Err(err).Msg("Failed to get databases when load register agent")
		return nil, err
	} else {
		registerAgent.Databases = s.DbApi.ToProtoDatabases(dbs)
		return registerAgent, nil
	}
}

type registerAgentLoader struct {
	agent         *GrpcAgentServer
	registerAgent *proto.RegisterInstance
}

func (r *registerAgentLoader) Run() bool {
	registerAgent, err := r.agent.loadRegisterAgent()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to register service")
		return false
	}
	r.registerAgent = registerAgent
	return true
}

type grpcServiceLoader struct {
	agent         *GrpcAgentServer
	registerAgent *proto.RegisterInstance
	service       proto.DbJobSvc_RegisterClient
}

func (g *grpcServiceLoader) Run() bool {
	service, err := g.agent.GrpcClient.Register(g.agent.QuitCtx, g.registerAgent)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to register service")
		return false
	}
	g.service = service
	return true
}
