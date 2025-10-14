package utils

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestLogger_VerboseGatesDebug(t *testing.T) {
	// Use a buffer-backed logger for deterministic assertions
	buf := &bytes.Buffer{}
	std = &Logger{verbose: false, logger: log.New(buf, "", 0)}

	Debug("should not appear: %s", "hidden")
	Info("info: %s", "hello")
	Warn("warn: %s", "there")
	Error("err: %s", "boom")

	out := buf.String()
	if strings.Contains(out, "should not appear") {
		t.Fatalf("debug output should be suppressed when verbose=false: %q", out)
	}
	if !strings.Contains(out, "INFO: info: hello") {
		t.Fatalf("expected info to be logged; got: %q", out)
	}
	if !strings.Contains(out, "WARN: warn: there") || !strings.Contains(out, "ERROR: err: boom") {
		t.Fatalf("expected warn/error to be logged; got: %q", out)
	}

	// Now enable verbose and expect debug to appear
	buf.Reset()
	SetVerbose(true)
	Debug("now appears: %s", "debug")
	out = buf.String()
	if !strings.Contains(out, "DEBUG: now appears: debug") {
		t.Fatalf("expected debug output when verbose=true, got: %q", out)
	}
}
