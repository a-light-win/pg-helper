package agent

import (
	"context"
	"time"

	"github.com/a-light-win/pg-helper/api/proto"
	"github.com/a-light-win/pg-helper/internal/db"
	"github.com/a-light-win/pg-helper/internal/utils"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

func (a *Agent) initGrpc() error {
	a.GrpcCallCtx, a.CancelGrpcCall = context.WithCancel(context.Background())

	dialOptions := []grpc.DialOption{}

	if tlsConfig, err := a.Config.Grpc.TlsConfig(); err != nil {
		return err
	} else if tlsConfig != nil {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}

	ka := keepalive.ClientParameters{
		Time:                time.Second * 10,
		Timeout:             time.Second * 10,
		PermitWithoutStream: true,
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
		AgentId:   a.Config.Id,
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

func (a *Agent) runUntilSuccess(runer utils.Runner, wait time.Duration) bool {
	for {
		if runer.Run() {
			return true
		}

		select {
		case <-a.QuitCtx.Done():
			return false
		case <-a.GrpcCallCtx.Done():
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
	service, err := g.agent.GrpcClient.Register(g.agent.GrpcCallCtx, g.registerAgent)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to register service")
		return false
	}
	g.service = service
	return true
}

func (a *Agent) registerService() proto.DbTaskSvc_RegisterClient {
	wait := time.Second * 2

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
		task, err := service.Recv()
		if err != nil {
			log.Warn().Err(err).Msg("Failed to receive task")
			service = a.registerService()
			if service == nil {
				return
			}
			continue
		}

		if task == nil {
			continue
		}

		switch task.Task.(type) {
		case *proto.DbTask_CreateDatabase:
			// TODO: Create the database
		case *proto.DbTask_MigratedDatabase:
			// TODO: The database is already migrated to another pg instance
		}

		break
	}
}
