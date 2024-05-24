package server

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type CronElement struct {
	TriggerAt  time.Time
	HandleFunc CronHandleFunc
}

type CronHandleFunc func(triggerAt time.Time)

func (e *CronElement) GetName() string {
	return ""
}

type CronServer struct {
	Elements []*CronElement
	mutex    sync.Mutex

	wg                  sync.WaitGroup
	firstElementChanged chan struct{}
	exited              chan struct{}
}

func NewCronServer() *CronServer {
	cronServer := &CronServer{
		Elements:            make([]*CronElement, 0),
		firstElementChanged: make(chan struct{}),
		exited:              make(chan struct{}),
	}
	return cronServer
}

func (s *CronServer) Init(setter GlobalSetter) error {
	return nil
}

func (s *CronServer) PostInit(getter GlobalGetter) error {
	return nil
}

func (s *CronServer) Run() {
	log.Log().Msg("Cron server is running")
	for {
		if !s.waitForNextElement() {
			return
		}
		s.processArrivedElements()
	}
}

func (s *CronServer) Shutdown(ctx context.Context) {
	log.Log().Msg("Cron server is shutting down")

	close(s.exited)
	s.wg.Wait()

	log.Log().Msg("Cron server is down")
}

func (s *CronServer) Send(msg NamedElement) {
	element := msg.(*CronElement)

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.Elements) == 0 {
		s.Elements = append(s.Elements, element)
		s.firstElementChanged <- struct{}{}
		return
	}

	if element.TriggerAt.Before(s.Elements[0].TriggerAt) {
		s.Elements = append([]*CronElement{element}, s.Elements...)
		s.firstElementChanged <- struct{}{}
		return
	}

	for i, e := range s.Elements {
		if e.TriggerAt.After(element.TriggerAt) {
			s.Elements = append(s.Elements[:i], append([]*CronElement{element}, s.Elements[i:]...)...)
			return
		}
	}
	s.Elements = append(s.Elements, element)
}

func (s *CronServer) Producer() Producer {
	return s
}

func (s *CronServer) nextTriggerAt() time.Time {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.Elements) == 0 {
		return time.Now().Add(time.Hour)
	}
	return s.Elements[0].TriggerAt
}

func (s *CronServer) waitForNextElement() bool {
	triggerAt := s.nextTriggerAt()

	select {
	case <-s.exited:
		return false
	case <-s.firstElementChanged:
		return true
	case <-time.After(time.Until(triggerAt)):
		return true
	}
}

func (s *CronServer) processArrivedElements() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	for i := 0; i < len(s.Elements); i++ {
		if s.Elements[i].TriggerAt.Before(now) {
			go s.processElement(s.Elements[i])
			s.Elements = append(s.Elements[:i], s.Elements[i+1:]...)
			i--
		}
	}
}

func (s *CronServer) processElement(e *CronElement) {
	s.wg.Add(1)
	defer s.wg.Done()

	e.HandleFunc(e.TriggerAt)
}
