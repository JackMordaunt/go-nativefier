package nativefier

import (
	"os"
	"os/signal"
	"syscall"
)

var signals *SignalHandler

func init() {
	signals = NewSignalHandler()
}

// SignalHandler provides hooks to respond to OS signals.
type SignalHandler struct {
	terminateSignals chan os.Signal
	onTerminate      []func()
}

// NewSignalHandler returns an initialised SignalHandler
func NewSignalHandler() *SignalHandler {
	terminate := make(chan os.Signal)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)
	handler := &SignalHandler{
		terminateSignals: terminate,
		onTerminate:      []func(){},
	}
	go handler.Run()
	return handler
}

// Run selects watches signal channels for OS signals.
func (s *SignalHandler) Run() {
	<-s.terminateSignals
	for _, fn := range s.onTerminate {
		fn()
	}
	os.Exit(0)
}

// OnTerminate is called when SigTerm or SigInt is received.
func (s *SignalHandler) OnTerminate(fn func()) {
	if fn == nil {
		return
	}
	s.onTerminate = append(s.onTerminate, fn)
}
