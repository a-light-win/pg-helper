package handler

import (
	"context"
	"sync"

	"github.com/a-light-win/pg-helper/pkg/utils"
	"github.com/rs/zerolog/log"
)

type Producer interface {
	Send(msg any)
	Close()
}

type Consumer interface {
	Producer() Producer
	Server
}

type NamedElement interface {
	GetName() string
}

type BaseProducer[T NamedElement] struct {
	Elements chan T
	closed   utils.AtomicBool
}

func (p *BaseProducer[T]) Send(msg any) {
	if element, ok := msg.(T); ok {
		if !p.closed.Get() {
			p.Elements <- element
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
		close(p.Elements)
	}
}

type BaseConsumer[T NamedElement] struct {
	Name string

	Elements chan T
	Handler  Handler

	MaxConcurrency int
	wg             sync.WaitGroup
}

func NewBaseConsumer[T NamedElement](name string, handler Handler, maxConcurrency int) *BaseConsumer[T] {
	return &BaseConsumer[T]{
		Name:           name,
		Elements:       make(chan T),
		Handler:        handler,
		MaxConcurrency: max(maxConcurrency, 1),
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
}

func (c *BaseConsumer[T]) Shutdown(ctx context.Context) error {
	log.Log().Str("Name", c.Name).Msg("Consumer is waiting for gracefule shutdown")

	c.wg.Wait()

	log.Log().Str("Name", c.Name).Msg("Consumer is shutdown gracefully")

	return nil
}

func (c *BaseConsumer[T]) Init(setter GlobalSetter) error {
	setter.Set(c.Name+"_producer", c.Producer())

	return nil
}

func (c *BaseConsumer[T]) PostInit(getter GlobalGetter) error {
	return nil
}
