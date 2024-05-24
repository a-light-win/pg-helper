package server

import "context"

type Server interface {
	Initialize
	Runner
	Shutdowner
}

type Producer interface {
	Send(msg NamedElement)
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

type Initialize interface {
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
}

type InitializableHandler interface {
	Handler
	Initialize
}

type FileChangedHandler interface {
	InitializableHandler

	FilesToWatch() []string
	OnWatchError(error)
}
