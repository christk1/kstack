package postgres

import (
	"os"
	"testing"
)

func TestPostgresAddon_Basics(t *testing.T) {
	a := &postgresAddon{}
	if a.Name() != "postgres" {
		t.Fatalf("unexpected name: %s", a.Name())
	}
	if a.Chart() == "" || a.RepoName() == "" || a.RepoURL() == "" {
		t.Fatalf("chart/repo fields should be set")
	}
	if a.Namespace() != "postgres" {
		t.Fatalf("unexpected namespace: %s", a.Namespace())
	}
}

func TestPostgresAddon_ValuesFiles_TempFile(t *testing.T) {
	a := &postgresAddon{}
	files := a.ValuesFiles()
	if len(files) != 1 {
		t.Fatalf("expected 1 values file, got %d", len(files))
	}
	if _, err := os.Stat(files[0]); err != nil {
		t.Fatalf("temp values file does not exist: %v", err)
	}
}

func TestHAValuesFile_ReturnsTempFile(t *testing.T) {
	p, err := HAValuesFile()
	if err != nil {
		t.Fatalf("HAValuesFile error: %v", err)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatalf("ha values file missing: %v", err)
	}
}
