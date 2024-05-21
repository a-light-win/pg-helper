package handler

type Handler interface {
	Handle(msg any) error
}
