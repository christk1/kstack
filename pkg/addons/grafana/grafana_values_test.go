package grafana_test

import (
	"os"
	"testing"

	"github.com/christk1/kstack/pkg/addons"
	_ "github.com/christk1/kstack/pkg/addons/grafana"
)

func TestGrafanaValuesInjection(t *testing.T) {
	a, err := addons.Get("grafana")
	if err != nil {
		t.Fatalf("grafana addon not registered: %v", err)
	}
	files := a.ValuesFiles()
	if len(files) == 0 {
		t.Fatalf("expected at least one values file, got none")
	}
	b, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("read values file failed: %v", err)
	}
	s := string(b)
	if !contains(s, "http://prometheus-server") {
		t.Fatalf("expected prometheus datasource in values, not found")
	}
	if !contains(s, "Example App: Requests") {
		t.Fatalf("expected dashboard title in injected values, not found")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || (len(sub) > 0 && (indexOf(s, sub) >= 0)))
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
