package cluster

import (
	"context"
	"testing"
	"time"
)

func TestKindProvider_DryRunPaths(t *testing.T) {
	p := NewKindProvider("gc-test")
	SetDryRun(p, true)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := p.Create(ctx); err != nil {
		t.Fatalf("Create dry-run failed: %v", err)
	}
	if err := p.Delete(ctx); err != nil {
		t.Fatalf("Delete dry-run failed: %v", err)
	}
	if exists, err := p.Exists(ctx); err != nil {
		t.Fatalf("Exists dry-run error: %v", err)
	} else if exists {
		t.Fatalf("Exists dry-run should default to false")
	}
	if path, err := p.KubeconfigPath(ctx); err != nil {
		t.Fatalf("KubeconfigPath dry-run error: %v", err)
	} else if path != "" {
		t.Fatalf("KubeconfigPath dry-run should be empty, got %q", path)
	}
}

func TestK3dProvider_DryRunPaths(t *testing.T) {
	p := NewK3dProvider("gc-test")
	SetDryRun(p, true)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := p.Create(ctx); err != nil {
		t.Fatalf("Create dry-run failed: %v", err)
	}
	if err := p.Delete(ctx); err != nil {
		t.Fatalf("Delete dry-run failed: %v", err)
	}
	if exists, err := p.Exists(ctx); err != nil {
		t.Fatalf("Exists dry-run error: %v", err)
	} else if exists {
		t.Fatalf("Exists dry-run should default to false")
	}
	if path, err := p.KubeconfigPath(ctx); err != nil {
		t.Fatalf("KubeconfigPath dry-run error: %v", err)
	} else if path != "" {
		t.Fatalf("KubeconfigPath dry-run should be empty, got %q", path)
	}
}
