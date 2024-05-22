package handler

import (
	"context"

	"github.com/rs/zerolog/log"
)

type Server interface {
	Initialization
	Runner
	Shutdowner
}

type Runner interface {
	Run()
}

type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

type Initialization interface {
	Init(setter GlobalSetter) error
	PostInit(getter GlobalGetter) error
}

type GlobalSetter interface {
	Set(key string, value interface{})
}

type GlobalGetter interface {
	Get(key string) interface{}
}

type BaseServer struct {
	Name    string
	Servers []Server

	QuitCtx context.Context
	Quit    context.CancelFunc

	globalVars map[string]interface{}
}

func (s *BaseServer) Set(key string, value interface{}) {
	if s.globalVars == nil {
		s.globalVars = make(map[string]interface{})
	}
	s.globalVars[key] = value
}

func (s *BaseServer) Get(key string) interface{} {
	return s.globalVars[key]
}

func (s *BaseServer) Init() error {
	for _, server := range s.Servers {
		if err := server.Init(s); err != nil {
			return err
		}
	}

	for _, server := range s.Servers {
		if err := server.PostInit(s); err != nil {
			return err
		}
	}

	return nil
}

func (s *BaseServer) Run() {
	for _, server := range s.Servers {
		go server.Run()
	}

	log.Log().Msgf("%s is running", s.Name)

	<-s.QuitCtx.Done()
	s.Quit()

	s.Shutdown()
}

func (s *BaseServer) Shutdown() {
	waitExitCtx := context.Background()

	for i := len(s.Servers) - 1; i >= 0; i-- {
		s.Servers[i].Shutdown(waitExitCtx)
	}

	log.Log().Msgf("%s is shutting down.", s.Name)
}
