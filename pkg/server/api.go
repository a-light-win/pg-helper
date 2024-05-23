package server

import "context"

type Server interface {
	Initialization
	Runner
	Shutdowner
}

type Producer interface {
	Send(msg NamedElement)
	Close()
}

type Consumer interface {
	Producer() Producer
	Server
}

type NamedElement interface {
	GetName() string
}

type Runner interface {
	Run()
}

type Shutdowner interface {
	Shutdown(ctx context.Context)
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

type Handler interface {
	Handle(msg NamedElement) error

	Initialization
}

type FileChangedHandler interface {
	Handler

	FilesToWatch() []string
	OnWatchError(error)
}
