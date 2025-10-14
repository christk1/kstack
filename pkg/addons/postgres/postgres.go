package postgres

import (
	"embed"
	"os"

	"github.com/christk1/kstack/pkg/addons"
)

//go:embed values.yaml values-ha.yaml
var valuesFS embed.FS

type postgresAddon struct{}

func (p *postgresAddon) Name() string      { return "postgres" }
func (p *postgresAddon) Chart() string     { return "bitnami/postgresql" }
func (p *postgresAddon) RepoName() string  { return "bitnami" }
func (p *postgresAddon) RepoURL() string   { return "https://charts.bitnami.com/bitnami" }
func (p *postgresAddon) Namespace() string { return "postgres" }
func (p *postgresAddon) ValuesFiles() []string {
	b, _ := valuesFS.ReadFile("values.yaml")
	f, _ := os.CreateTemp("", "kstack-postgres-values-")
	if f != nil {
		f.Write(b)
		f.Close()
		return []string{f.Name()}
	}
	return nil
}

// HAValuesFile returns a temporary file path containing the HA chart values
// to be used with the Bitnami postgresql-ha chart. The caller is responsible
// for removing the returned file when no longer needed.
func HAValuesFile() (string, error) {
	b, err := valuesFS.ReadFile("values-ha.yaml")
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp("", "kstack-postgres-ha-values-")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(b); err != nil {
		f.Close()
		return "", err
	}
	if err := f.Close(); err != nil {
		return "", err
	}
	return f.Name(), nil
}

func init() {
	addons.Register(&postgresAddon{})
}
