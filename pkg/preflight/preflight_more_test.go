package preflight

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func writeFake(t *testing.T, name, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("skip on windows")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake %s: %v", name, err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return p
}

func TestCheckProviderCLI_Success(t *testing.T) {
	writeFake(t, "kind", "#!/usr/bin/env bash\nexit 0\n")
	if err := CheckProviderCLI("kind"); err != nil {
		t.Fatalf("expected success for kind in PATH: %v", err)
	}
	writeFake(t, "k3d", "#!/usr/bin/env bash\nexit 0\n")
	if err := CheckProviderCLI("k3d"); err != nil {
		t.Fatalf("expected success for k3d in PATH: %v", err)
	}
}

func TestCheckDocker_Success_Verbose(t *testing.T) {
	writeFake(t, "docker", "#!/usr/bin/env bash\nif [ \"$1\" = \"info\" ]; then echo ok; exit 0; fi\nexit 0\n")
	oldV, oldD := Verbose, Debug
	Verbose, Debug = true, false
	defer func() { Verbose, Debug = oldV, oldD }()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := CheckDocker(ctx); err != nil {
		t.Fatalf("expected docker success, got %v", err)
	}
}

func TestCheckDocker_PermissionDenied_Error_NoVerbose(t *testing.T) {
	writeFake(t, "docker", "#!/usr/bin/env bash\nif [ \"$1\" = \"info\" ]; then echo permission denied >&2; exit 1; fi\nexit 0\n")
	oldV, oldD := Verbose, Debug
	Verbose, Debug = false, false
	defer func() { Verbose, Debug = oldV, oldD }()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := CheckDocker(ctx)
	if err == nil {
		t.Fatalf("expected error")
	}
	msg := err.Error()
	if !strings.Contains(strings.ToLower(msg), "permission denied") {
		t.Fatalf("expected permission denied in error, got: %s", msg)
	}
	if strings.Contains(msg, "docker info output") {
		t.Fatalf("did not expect full docker output in non-verbose mode")
	}
}

func TestCheckDocker_CannotConnect_Debug_ShowsOutput(t *testing.T) {
	writeFake(t, "docker", "#!/usr/bin/env bash\nif [ \"$1\" = \"info\" ]; then echo cannot connect to the Docker daemon >&2; exit 1; fi\nexit 0\n")
	oldV, oldD := Verbose, Debug
	Verbose, Debug = false, true
	defer func() { Verbose, Debug = oldV, oldD }()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := CheckDocker(ctx)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "docker output") {
		t.Fatalf("expected docker output section in error, got: %s", err)
	}
}
