package handler

import "context"

type Server interface {
	Initialization
	Run()
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
