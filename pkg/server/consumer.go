package server

import (
	"context"
	"sync"

	"github.com/a-light-win/pg-helper/pkg/utils"
	"github.com/a-light-win/pg-helper/pkg/utils/logger"
	"github.com/rs/zerolog/log"
)

type BaseConsumer[T NamedElement] struct {
	Name string

	Elements chan T
	Handler  InitializableHandler

	MaxConcurrency int
	wg             sync.WaitGroup
	exited         chan struct{}

	chanClosed utils.AtomicBool
	addingWg   sync.WaitGroup
}

func NewBaseConsumer[T NamedElement](name string, handler InitializableHandler, maxConcurrency int) *BaseConsumer[T] {
	return &BaseConsumer[T]{
		Name:           name,
		Elements:       make(chan T),
		Handler:        handler,
		MaxConcurrency: max(maxConcurrency, 1),
		exited:         make(chan struct{}),
	}
}

func (c *BaseConsumer[T]) Producer() Producer {
	return c
}

func (c *BaseConsumer[T]) Run() {
	log.Log().Msgf("%s is running", c.Name)

	sem := make(chan struct{}, c.MaxConcurrency)
	defer close(sem)

	for {
		element, ok := <-c.Elements
		if !ok {
			break
		}

		c.wg.Add(1)
		sem <- struct{}{}
		go func() {
			if err := c.Handler.Handle(element); err != nil {
				if _, ok := err.(*logger.AlreadyLoggedError); !ok {
					log.Warn().
						Err(err).
						Str("Name", element.GetName()).
						Str("Consumer", c.Name).
						Msg("Failed to handle element")
				}
			}
			<-sem
			c.wg.Done()
		}()
	}

	c.exited <- struct{}{}
}

func (c *BaseConsumer[T]) Shutdown(ctx context.Context) {
	log.Log().Msgf("%s is shutting down", c.Name)

	c.Close()
	<-c.exited
	c.wg.Wait()

	if shutdowner, ok := c.Handler.(Shutdowner); ok {
		shutdowner.Shutdown(ctx)
	}

	log.Log().Msgf("%s is down", c.Name)
}

func (c *BaseConsumer[T]) Init(setter GlobalSetter) error {
	if err := c.Handler.Init(setter); err != nil {
		return err
	}

	return nil
}

func (c *BaseConsumer[T]) PostInit(getter GlobalGetter) error {
	if err := c.Handler.PostInit(getter); err != nil {
		return err
	}

	return nil
}

func (c *BaseConsumer[T]) Send(msg NamedElement) {
	if element, ok := msg.(T); ok {
		if !c.chanClosed.Get() {
			c.addingWg.Add(1)
			c.Elements <- element
			c.addingWg.Done()
		} else {
			log.Warn().Str("Name", element.GetName()).
				Msg("Producer is closed, discard element")
		}
	} else {
		log.Error().
			Str("MsgName", msg.GetName()).
			Str("Server", c.Name).
			Msg("Invalid element type")
	}
}

func (c *BaseConsumer[T]) Close() {
	if c.chanClosed.CompareAndSwap(false, true) {
		c.addingWg.Wait()
		close(c.Elements)
	}
}
