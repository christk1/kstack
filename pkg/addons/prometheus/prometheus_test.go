package prometheus

import (
	"testing"
)

func TestPrometheusAddon_Basics(t *testing.T) {
	a := &prometheusAddon{}
	if a.Name() != "prometheus" || a.Chart() == "" || a.RepoName() == "" || a.RepoURL() == "" {
		t.Fatalf("prometheus addon fields invalid")
	}
	if a.Namespace() != "monitoring" {
		t.Fatalf("unexpected ns: %s", a.Namespace())
	}
	// Values file may or may not exist depending on the repo layout; just call it for coverage.
	_ = a.ValuesFiles()
}
