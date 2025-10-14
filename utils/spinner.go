package utils

import (
	"fmt"
	"os"
	"time"
)

// Spinner is a very small terminal spinner for progress feedback.
// It writes to stderr to avoid mixing with command outputs printed to stdout.
type Spinner struct {
	msg     string
	stopCh  chan struct{}
	running bool
}

// NewSpinner creates a new spinner with a message prefix.
func NewSpinner(msg string) *Spinner { return &Spinner{msg: msg, stopCh: make(chan struct{})} }

// Start begins the spinner animation. If already running, it's a no-op.
func (s *Spinner) Start() {
	if s.running {
		return
	}
	s.running = true
	go func() {
		frames := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
		i := 0
		for {
			select {
			case <-s.stopCh:
				// clear line
				fmt.Fprint(os.Stderr, "\r\x1b[2K")
				return
			default:
				fmt.Fprintf(os.Stderr, "\r%s %c", s.msg, frames[i%len(frames)])
				time.Sleep(90 * time.Millisecond)
				i++
			}
		}
	}()
}

// Stop stops the spinner and clears the line.
func (s *Spinner) Stop() {
	if !s.running {
		return
	}
	close(s.stopCh)
	s.running = false
}
