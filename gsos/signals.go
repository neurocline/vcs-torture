// vcs-torture/gsos/signals.go

package gsos

import (
	"os"
	"os/signal"
	"syscall"
)

// Catch signals so that user can't accidentally interrupt something
// sensitive (like writing a log file) - e.g. clean abort
// For now, just catches SIGINT (ctrl-c)
type CatchSignals struct {
	Abort bool

	sigs   chan os.Signal
	cancel bool
}

// Start capturing signals
func (s *CatchSignals) Capture() {
	s.sigs = make(chan os.Signal, 1)
	s.Abort = false
	s.cancel = false

	signal.Notify(s.sigs, syscall.SIGINT, syscall.SIGTERM)
	go func(s *CatchSignals) {
		_ = <-s.sigs // change this if we care about which signal
		if !s.cancel {
			s.Abort = true
		}
	}(s)
}

// Stop capturing signals
func (s *CatchSignals) Release() {
	s.cancel = true
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)
	close(s.sigs)
}
