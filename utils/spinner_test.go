package utils

import (
	"testing"
	"time"
)

func TestSpinner_StartStop_Idempotent(t *testing.T) {
	s := NewSpinner("testing")
	s.Start()
	// calling Start again should be a no-op
	s.Start()
	time.Sleep(50 * time.Millisecond)
	s.Stop()
	// calling Stop again should be a no-op
	s.Stop()
}
