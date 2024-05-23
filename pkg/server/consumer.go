package server

import (
	"context"
	"sync"

	"github.com/a-light-win/pg-helper/pkg/utils"
	"github.com/rs/zerolog/log"
)

type BaseConsumer[T NamedElement] struct {
	Name string

	Elements chan T
	Handler  Handler

	MaxConcurrency int
	wg             sync.WaitGroup
	exited         chan struct{}
}

func NewBaseConsumer[T NamedElement](name string, handler Handler, maxConcurrency int) *BaseConsumer[T] {
	return &BaseConsumer[T]{
		Name:           name,
		Elements:       make(chan T),
		Handler:        handler,
		MaxConcurrency: max(maxConcurrency, 1),
		exited:         make(chan struct{}),
	}
}

func (c *BaseConsumer[T]) Producer() Producer {
	return &BaseProducer[T]{Elements: c.Elements}
}

func (c *BaseConsumer[T]) Run() {
	log.Log().Msgf("%s is running", c.Name)

	sem := make(chan struct{}, c.MaxConcurrency)
	defer close(sem)

	for {
		element, ok := <-c.Elements
		if !ok {
			log.Log().Str("Name", c.Name).
				Msg("Consumer is closed")
			break
		}

		c.wg.Add(1)
		sem <- struct{}{}
		go func() {
			if err := c.Handler.Handle(element); err != nil {
				log.Warn().
					Err(err).
					Str("Name", element.GetName()).
					Str("Consumer", c.Name).
					Msg("Failed to handle element")
			}
			<-sem
			c.wg.Done()
		}()
	}

	c.exited <- struct{}{}
}

func (c *BaseConsumer[T]) Shutdown(ctx context.Context) {
	log.Log().Str("Name", c.Name).Msg("Consumer is waiting for gracefule shutdown")

	close(c.Elements)
	<-c.exited
	c.wg.Wait()

	if shutdowner, ok := c.Handler.(Shutdowner); ok {
		shutdowner.Shutdown(ctx)
	}

	log.Log().Str("Name", c.Name).Msg("Consumer is shutdown gracefully")
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

type BaseProducer[T NamedElement] struct {
	Elements chan T

	closed utils.AtomicBool
	wg     sync.WaitGroup
}

func (p *BaseProducer[T]) Send(msg NamedElement) {
	if element, ok := msg.(T); ok {
		if !p.closed.Get() {
			p.wg.Add(1)
			p.Elements <- element
			p.wg.Done()
		} else {
			log.Warn().Str("Name", element.GetName()).
				Msg("Producer is closed, discard element")
		}
	} else {
		log.Error().Interface("Element", msg).
			Msg("Invalid element type")
	}
}

func (p *BaseProducer[T]) Close() {
	if p.closed.CompareAndSwap(false, true) {
		p.wg.Wait()
		close(p.Elements)
	}
}
