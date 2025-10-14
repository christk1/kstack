package prometheus

import (
	"os"

	"github.com/christk1/kstack/pkg/addons"
)

type prometheusAddon struct{}

func (p *prometheusAddon) Name() string     { return "prometheus" }
func (p *prometheusAddon) Chart() string    { return "prometheus-community/prometheus" }
func (p *prometheusAddon) RepoName() string { return "prometheus-community" }
func (p *prometheusAddon) RepoURL() string {
	return "https://prometheus-community.github.io/helm-charts"
}
func (p *prometheusAddon) Namespace() string { return "monitoring" }
func (p *prometheusAddon) ValuesFiles() []string {
	pth := "pkg/addons/prometheus/values.yaml"
	if _, err := os.Stat(pth); err == nil {
		return []string{pth}
	}
	return nil
}

func init() {
	addons.Register(&prometheusAddon{})
}
