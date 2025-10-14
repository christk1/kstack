package helm

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func writeFailHelm(t *testing.T, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("skip on windows")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "helm")
	if err := os.WriteFile(p, []byte(script), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func TestHelmErrors_Branches(t *testing.T) {
	// version fails
	h := NewClient(writeFailHelm(t, "#!/usr/bin/env bash\nexit 1\n"))
	if _, err := h.Preflight(time.Second); err == nil {
		t.Fatalf("expected preflight error")
	}

	// repo add fails
	h = NewClient(writeFailHelm(t, "#!/usr/bin/env bash\nif [ \"$1\" = \"repo\" ]; then exit 1; fi\necho ok\n"))
	if err := h.RepoAdd("n", "u"); err == nil {
		t.Fatalf("expected repo add error")
	}

	// repo update fails
	h = NewClient(writeFailHelm(t, "#!/usr/bin/env bash\nif [ \"$1\" = \"repo\" ] && [ \"$2\" = \"update\" ]; then exit 1; fi\necho ok\n"))
	if err := h.RepoUpdate(); err == nil {
		t.Fatalf("expected repo update error")
	}

	// install fails
	h = NewClient(writeFailHelm(t, "#!/usr/bin/env bash\nif [ \"$1\" = \"upgrade\" ]; then exit 1; fi\necho ok\n"))
	if err := h.InstallOrUpgrade("r", "c", "ns", "", false, time.Second, false, nil); err == nil {
		t.Fatalf("expected install error")
	}

	// uninstall fails
	h = NewClient(writeFailHelm(t, "#!/usr/bin/env bash\nif [ \"$1\" = \"uninstall\" ]; then exit 1; fi\necho ok\n"))
	if err := h.Uninstall("r", "ns", false, time.Second); err == nil {
		t.Fatalf("expected uninstall error")
	}

	// list returns invalid json
	h = NewClient(writeFailHelm(t, "#!/usr/bin/env bash\nif [ \"$1\" = \"list\" ]; then echo not-json; exit 0; fi\necho ok\n"))
	if _, err := h.ListReleases("ns"); err == nil {
		t.Fatalf("expected list parse error")
	}
}
