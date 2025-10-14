package kafka

import (
	"os"
	"testing"
)

func TestKafkaAddon_Basics(t *testing.T) {
	a := &kafkaAddon{}
	if a.Name() != "kafka" || a.Chart() == "" || a.RepoName() == "" || a.RepoURL() == "" {
		t.Fatalf("kafka addon fields invalid")
	}
	if a.Namespace() != "kafka" {
		t.Fatalf("unexpected namespace: %s", a.Namespace())
	}
}

func TestKafkaAddon_ValuesTempFile(t *testing.T) {
	files := (&kafkaAddon{}).ValuesFiles()
	if len(files) != 1 {
		t.Fatalf("expected 1 values file, got %d", len(files))
	}
	if _, err := os.Stat(files[0]); err != nil {
		t.Fatalf("temp values file missing: %v", err)
	}
}
