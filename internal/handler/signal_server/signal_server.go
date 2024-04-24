package signal_server

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

type SignalServer struct {
	Quit    context.CancelFunc
	QuitCtx context.Context
}

func NewSignalServer() *SignalServer {
	ctx := context.Background()
	quitCtx, quit := context.WithCancel(ctx)
	return &SignalServer{Quit: quit, QuitCtx: quitCtx}
}

func (s *SignalServer) Run() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-signalChan:
		s.Quit()
	case <-s.QuitCtx.Done():
	}
}

func (s *SignalServer) Shutdown(ctx context.Context) error {
	s.Quit()
	return nil
}
