package cluster

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func writeFakeCmd(t *testing.T, name string, content string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o755); err != nil {
		t.Fatalf("write fake %s: %v", name, err)
	}
	// prepend to PATH
	os.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	return path
}

func TestKindProvider_NonDryRun_WithFakeKind(t *testing.T) {
	script := "#!/usr/bin/env bash\nset -e\nif [ \"$1\" = \"--version\" ]; then echo kind v0.0.0; exit 0; fi\ncase \"$1 $2\" in\n  'create cluster') echo ok;;\n  'delete cluster') echo ok;;\n  'get clusters') echo 'gc-test';;\n  'get kubeconfig') echo 'apiVersion: v1';;\n  *) echo ok;;\n esac\n"
	writeFakeCmd(t, "kind", script)
	p := NewKindProvider("gc-test").(*kindProvider)
	p.DryRun = false
	// also fake docker for preflight
	writeFakeCmd(t, "docker", "#!/usr/bin/env bash\nexit 0\n")

	// Call Exists (should be true), KubeconfigPath, Create and Delete to exercise real exec paths.
	if ok, err := p.Exists(context.Background()); err != nil || !ok {
		t.Fatalf("exists expected true, err=%v", err)
	}
	if path, err := p.KubeconfigPath(context.Background()); err != nil || path == "" {
		t.Fatalf("expected kubeconfig path, err=%v path=%q", err, path)
	} else {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("kubeconfig file missing: %v", err)
		}
	}
	if err := p.Create(context.Background()); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if err := p.Delete(context.Background()); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
}

func TestK3dProvider_NonDryRun_WithFakeK3d(t *testing.T) {
	script := "#!/usr/bin/env bash\nset -e\nif [ \"$1\" = \"version\" ]; then echo k3d v0.0.0; exit 0; fi\nif [ \"$1\" = \"kubeconfig\" ] && [ \"$2\" = \"get\" ]; then echo 'apiVersion: v1'; exit 0; fi\nif [ \"$1\" = \"cluster\" ] && [ \"$2\" = \"list\" ]; then echo 'gc-test'; exit 0; fi\nif [ \"$1\" = \"cluster\" ] && [ \"$2\" = \"create\" ]; then exit 0; fi\nif [ \"$1\" = \"cluster\" ] && [ \"$2\" = \"delete\" ]; then exit 0; fi\nexit 0\n"
	writeFakeCmd(t, "k3d", script)
	writeFakeCmd(t, "docker", "#!/usr/bin/env bash\nexit 0\n")
	p := NewK3dProvider("gc-test").(*k3dProvider)
	p.DryRun = false

	if ok, err := p.Exists(context.Background()); err != nil || !ok {
		t.Fatalf("exists expected true, err=%v", err)
	}
	if path, err := p.KubeconfigPath(context.Background()); err != nil || path == "" {
		t.Fatalf("expected kubeconfig path, err=%v path=%q", err, path)
	}
	if err := p.Create(context.Background()); err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if err := p.Delete(context.Background()); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
}
