package cluster

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func writeFailCmd(t *testing.T, name, script string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("skip on windows")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(script), 0o755); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	os.Setenv("PATH", dir+":"+os.Getenv("PATH"))
}

func TestKindProvider_ErrorBranches(t *testing.T) {
	// kind get clusters fails
	writeFailCmd(t, "kind", "#!/usr/bin/env bash\nexit 1\n")
	p := NewKindProvider("gc-x").(*kindProvider)
	p.DryRun = false
	if _, err := p.Exists(context.Background()); err == nil {
		t.Fatalf("expected exists error")
	}

	// kubeconfig fails
	writeFailCmd(t, "kind", "#!/usr/bin/env bash\nif [ \"$1\" = \"get\" ] && [ \"$2\" = \"kubeconfig\" ]; then exit 1; fi\necho ok\n")
	if _, err := p.KubeconfigPath(context.Background()); err == nil {
		t.Fatalf("expected kubeconfig error")
	}

	// create fails (preflight ok but create returns non-zero)
	writeFailCmd(t, "kind", "#!/usr/bin/env bash\nif [ \"$1\" = \"--version\" ]; then exit 0; fi\nif [ \"$1\" = \"create\" ]; then exit 1; fi\necho ok\n")
	writeFailCmd(t, "docker", "#!/usr/bin/env bash\nif [ \"$1\" = \"info\" ]; then exit 0; fi\nexit 0\n")
	if err := p.Create(context.Background()); err == nil {
		t.Fatalf("expected create error")
	}

	// delete fails
	writeFailCmd(t, "kind", "#!/usr/bin/env bash\nif [ \"$1\" = \"delete\" ]; then exit 1; fi\necho ok\n")
	if err := p.Delete(context.Background()); err == nil {
		t.Fatalf("expected delete error")
	}
}

func TestK3dProvider_ErrorBranches(t *testing.T) {
	// Exists failure
	writeFailCmd(t, "k3d", "#!/usr/bin/env bash\nexit 1\n")
	p := NewK3dProvider("gc-y").(*k3dProvider)
	p.DryRun = false
	if _, err := p.Exists(context.Background()); err == nil {
		t.Fatalf("expected exists error")
	}

	// Kubeconfig fails
	writeFailCmd(t, "k3d", "#!/usr/bin/env bash\nif [ \"$1\" = \"kubeconfig\" ] && [ \"$2\" = \"get\" ]; then exit 1; fi\necho ok\n")
	if _, err := p.KubeconfigPath(context.Background()); err == nil {
		t.Fatalf("expected kubeconfig error")
	}

	// Create fails
	writeFailCmd(t, "k3d", "#!/usr/bin/env bash\nif [ \"$1\" = \"cluster\" ] && [ \"$2\" = \"create\" ]; then exit 1; fi\necho ok\n")
	if err := p.Create(context.Background()); err == nil {
		t.Fatalf("expected create error")
	}

	// Delete fails
	writeFailCmd(t, "k3d", "#!/usr/bin/env bash\nif [ \"$1\" = \"cluster\" ] && [ \"$2\" = \"delete\" ]; then exit 1; fi\necho ok\n")
	if err := p.Delete(context.Background()); err == nil {
		t.Fatalf("expected delete error")
	}
}
