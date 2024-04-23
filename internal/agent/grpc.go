package agent

import (
	"context"
	"errors"
	"io"
	"math/rand"
	"time"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/agent/grpc_handler"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
)

func (a *Agent) initGrpc() error {
	dialOptions := []grpc.DialOption{}

	creds, err := a.Config.Grpc.Tls.Credentials()
	if err != nil {
		return err
	}
	dialOptions = append(dialOptions, grpc.WithTransportCredentials(creds))

	if a.Config.Grpc.ServerName != "" {
		dialOptions = append(dialOptions, grpc.WithAuthority(a.Config.Grpc.ServerName))
	}

	authToken, err := a.Config.Grpc.AuthToken()
	if err == nil {
		log.Log().Msg("Auth token is provided")
		authCreds := grpc_handler.NewAuthToken(authToken, a.Config.Grpc.Tls.Enabled)
		dialOptions = append(dialOptions, grpc.WithPerRPCCredentials(authCreds))
	} else if !a.Config.Grpc.Tls.MTLSEnabled {
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

	conn, err := grpc.Dial(a.Config.Grpc.Url, dialOptions...)
	if err != nil {
		log.Error().Err(err).Msg("Failed to dial gRPC server")
		return err
	}

	a.GrpcClient = proto.NewDbTaskSvcClient(conn)
	return nil
}

func (a *Agent) loadRegisterAgent() (*proto.RegisterAgent, error) {
	registerAgent := &proto.RegisterAgent{
		PgVersion: a.Config.Db.CurrentVersion,
	}

	conn, err := a.DbPool.Acquire(context.Background())
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire connection when load register agent")
		return nil, err
	}
	defer conn.Release()

	q := db.New(conn)
	registerAgent.Databases, err = db.ListDatabases(q)
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire connection when load register agent")
		return nil, err
	}
	return registerAgent, nil
}

func (a *Agent) runUntilSuccess(runer utils.Runner, firstWait int) bool {
	continueWait := firstWait
	maxContinueWait := 60

	for {

		select {
		case <-a.QuitCtx.Done():
			return false
		case <-a.JobCtx.Done():
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
		case <-a.QuitCtx.Done():
			return false
		case <-a.JobCtx.Done():
			return false
		case <-time.After(wait):
			continue
		}
	}
}

type registerAgentLoader struct {
	agent         *Agent
	registerAgent *proto.RegisterAgent
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
	agent         *Agent
	registerAgent *proto.RegisterAgent
	service       proto.DbTaskSvc_RegisterClient
}

func (g *grpcServiceLoader) Run() bool {
	service, err := g.agent.GrpcClient.Register(g.agent.JobCtx, g.registerAgent)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to register service")
		return false
	}
	g.service = service
	return true
}

func (a *Agent) registerService() proto.DbTaskSvc_RegisterClient {
	wait := 2

	registerAgentLoader := &registerAgentLoader{agent: a}
	if !a.runUntilSuccess(registerAgentLoader, wait) {
		return nil
	}

	grpcServiceLoader := &grpcServiceLoader{agent: a, registerAgent: registerAgentLoader.registerAgent}
	if !a.runUntilSuccess(grpcServiceLoader, wait) {
		return nil
	}
	return grpcServiceLoader.service
}

func (a *Agent) runGrpc() {
	service := a.registerService()
	if service == nil {
		return
	}

	for {

		select {
		case <-a.QuitCtx.Done():
			return
		default:
		}

		task, err := service.Recv()
		if err != nil {
			s, ok := status.FromError(err)
			if ok && s.Code() == codes.Unauthenticated {
				log.Error().Err(err).Msg("Failed to receive task")
				return
			}

			if err == io.EOF {
				log.Info().Msg("Server closed the connection")
			} else {
				log.Warn().Err(err).Msg("Failed to receive task")
			}

			wait := time.Duration(rand.Intn(3)+1) * time.Second
			time.Sleep(wait)

			service = a.registerService()
			if service == nil {
				return
			}
			continue
		}

		if task == nil {
			continue
		}

		// TODO: support handle multiple tasks concurrently
		handler := grpc_handler.New(task)
		if err := handler.Validate(); err != nil {
			handler.OnError(err)
			continue
		}
		if err := handler.Handle(); err != nil {
			handler.OnError(err)
			continue
		}
	}
}
