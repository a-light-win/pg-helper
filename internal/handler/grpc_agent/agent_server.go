package grpc_agent

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"time"

	"github.com/a-light-win/pg-helper/api/proto"
	config "github.com/a-light-win/pg-helper/internal/config/agent"
	"github.com/a-light-win/pg-helper/internal/constants"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/job"
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/a-light-win/pg-helper/pkg/handler"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

type GrpcAgentServer struct {
	GrpcConfig *config.GrpcClientConfig
	GrpcClient proto.DbTaskSvcClient

	QuitCtx context.Context

	DbApi *db.DbApi

	handler *GrpcAgentHandler
}

func NewGrpcAgentServer(grpcConfig *config.GrpcClientConfig, quitCtx context.Context) *GrpcAgentServer {
	server := &GrpcAgentServer{
		GrpcConfig: grpcConfig,
		QuitCtx:    quitCtx,
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

	s.GrpcClient = proto.NewDbTaskSvcClient(conn)
	return nil
}

func (s *GrpcAgentServer) Init(setter handler.GlobalSetter) error {
	if err := s.initGrpcClient(); err != nil {
		return err
	}
	setter.Set(constants.AgentKeyGrpcClient, s.GrpcClient)

	return nil
}

func (s *GrpcAgentServer) PostInit(getter handler.GlobalGetter) error {
	s.DbApi = getter.Get(constants.AgentKeyDbApi).(*db.DbApi)
	jobProducer := getter.Get(constants.AgentKeyJobProducer).(*job.JobProducer)

	s.handler = NewGrpcAgentHandler(s.DbApi, s.GrpcClient, jobProducer)

	return nil
}

func (s *GrpcAgentServer) Run() {
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

		go s.handler.Run(task)
	}
}

func (s *GrpcAgentServer) Shutdown(ctx context.Context) error {
	// TODO: wait all task ready
	return nil
}

func (s *GrpcAgentServer) registerService() proto.DbTaskSvc_RegisterClient {
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

	err := s.DbApi.Query(func(q *db.Queries) error {
		var err error
		registerAgent.Databases, err = db.ListDatabases(q)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get databases when load register agent")
			return err
		}
		return nil
	})
	return registerAgent, err
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
	service       proto.DbTaskSvc_RegisterClient
}

func (g *grpcServiceLoader) Run() bool {
	service, err := g.agent.GrpcClient.Register(g.agent.DbApi.ConnCtx, g.registerAgent)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to register service")
		return false
	}
	g.service = service
	return true
}
