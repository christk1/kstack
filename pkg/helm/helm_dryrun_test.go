package helm

import (
	"testing"
	"time"
)

func TestHelmClient_DryRun_RepoAndInstall(t *testing.T) {
	h := NewClient("helm")
	h.DryRun = true
	if _, err := h.Preflight(2 * time.Second); err != nil {
		t.Fatalf("preflight dry-run failed: %v", err)
	}
	if err := h.RepoAdd("bitnami", "https://charts.bitnami.com/bitnami"); err != nil {
		t.Fatalf("repo add dry-run failed: %v", err)
	}
	if err := h.RepoUpdate(); err != nil {
		t.Fatalf("repo update dry-run failed: %v", err)
	}
	if err := h.InstallOrUpgrade("rel", "chart", "ns", "", true, 30*time.Second, true, []string{"k=v"}); err != nil {
		t.Fatalf("install dry-run failed: %v", err)
	}
	if err := h.Uninstall("rel", "ns", true, 10*time.Second); err != nil {
		t.Fatalf("uninstall dry-run failed: %v", err)
	}
	if _, err := h.ListReleases("ns"); err != nil {
		t.Fatalf("list releases dry-run failed: %v", err)
	}
}
