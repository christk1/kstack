package kafka

import (
	"embed"
	"os"

	"github.com/christk1/kstack/pkg/addons"
)

//go:embed values.yaml
var valuesFS embed.FS

type kafkaAddon struct{}

func (k *kafkaAddon) Name() string      { return "kafka" }
func (k *kafkaAddon) Chart() string     { return "bitnami/kafka" }
func (k *kafkaAddon) RepoName() string  { return "bitnami" }
func (k *kafkaAddon) RepoURL() string   { return "https://charts.bitnami.com/bitnami" }
func (k *kafkaAddon) Namespace() string { return "kafka" }
func (k *kafkaAddon) ValuesFiles() []string {
	b, _ := valuesFS.ReadFile("values.yaml")
	f, _ := os.CreateTemp("", "kstack-kafka-values-")
	if f != nil {
		f.Write(b)
		f.Close()
		return []string{f.Name()}
	}
	return nil
}

func init() {
	addons.Register(&kafkaAddon{})
}
