package exampleapp

import (
	"os"

	"github.com/christk1/kstack/pkg/addons"
)

// exampleAppAddon is a minimal addon that installs the local Helm chart at ./examples/app.
type exampleAppAddon struct{}

func (e *exampleAppAddon) Name() string      { return "example-app" }
func (e *exampleAppAddon) Chart() string     { return "./pkg/addons/exampleapp/chart" }
func (e *exampleAppAddon) RepoName() string  { return "" }
func (e *exampleAppAddon) RepoURL() string   { return "" }
func (e *exampleAppAddon) Namespace() string { return "app" }
func (e *exampleAppAddon) ValuesFiles() []string {
	// Prefer the addon chart's values if present
	p := "pkg/addons/exampleapp/chart/values.yaml"
	if _, err := os.Stat(p); err == nil {
		return []string{p}
	}
	return nil
}

func init() {
	addons.Register(&exampleAppAddon{})
}
