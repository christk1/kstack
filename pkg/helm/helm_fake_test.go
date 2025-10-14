package helm

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func writeFakeHelm(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "helm")
	script := "#!/usr/bin/env bash\nset -e\ncmd=$1; shift\ncase \"$cmd\" in\n  version) echo 'v3.12.0';;\n  repo) echo 'ok';;\n  upgrade) echo 'ok';;\n  uninstall) echo 'ok';;\n  list) echo '[]';;\n  *) echo 'ok';;\nesac\n"
	if runtime.GOOS == "windows" {
		t.Skip("fake helm script not supported on windows")
	}
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake helm: %v", err)
	}
	return path
}

func TestHelmClient_WithFakeBinary(t *testing.T) {
	h := NewClient(writeFakeHelm(t))
	if v, err := h.Preflight(2 * time.Second); err != nil || v == "" {
		t.Fatalf("preflight failed: %v, v=%q", err, v)
	}
	if err := h.RepoAdd("n", "u"); err != nil {
		t.Fatalf("repo add: %v", err)
	}
	if err := h.RepoUpdate(); err != nil {
		t.Fatalf("repo update: %v", err)
	}
	if err := h.InstallOrUpgrade("r", "c", "ns", "", true, time.Second, false, nil); err != nil {
		t.Fatalf("install: %v", err)
	}
	if err := h.Uninstall("r", "ns", true, time.Second); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if rels, err := h.ListReleases("ns"); err != nil || len(rels) != 0 {
		t.Fatalf("list: err=%v rels=%v", err, rels)
	}
}
