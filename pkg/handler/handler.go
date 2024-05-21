package handler

type Handler interface {
	Handle(msg NamedElement) error

	Initialization
}
