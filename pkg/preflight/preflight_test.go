package preflight

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCheckCommands_MissingAndPresent(t *testing.T) {
	// A definitely-missing command name
	missing := "kstack-definitely-missing-cmd-xyz"

	// Create a temporary executable to simulate a present command
	dir := t.TempDir()
	exeName := "presentcmd"
	path := filepath.Join(dir, exeName)
	script := "#!/bin/sh\nexit 0\n"
	if runtime.GOOS == "windows" {
		exeName += ".bat"
		path = filepath.Join(dir, exeName)
		script = "@echo off\nexit /B 0\n"
	}
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("failed to write temp executable: %v", err)
	}

	// Prepend temp dir to PATH so LookPath can find it
	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+oldPath)

	if err := CheckCommands([]string{missing, exeName}); err == nil {
		t.Fatalf("expected error for missing command, got nil")
	}

	if err := CheckCommands([]string{exeName}); err != nil {
		t.Fatalf("expected success for present command, got %v", err)
	}
}
