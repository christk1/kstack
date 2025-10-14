package cluster

import (
	"context"
)

// Provider defines the minimal operations required by a cluster implementation
// (e.g. kind, k3d). Implementations should be side-effecting and return
// actionable errors; callers should pass a context with timeouts.
type Provider interface {
	// Create creates the cluster. It should be idempotent where possible
	// (return a clear error if creation is impossible).
	Create(ctx context.Context) error

	// Delete tears down the cluster and releases any associated resources.
	Delete(ctx context.Context) error

	// Exists returns whether the cluster already exists.
	Exists(ctx context.Context) (bool, error)

	// KubeconfigPath returns the path to a kubeconfig file that can be used
	// to access the cluster. Providers may return an empty string if no
	// kubeconfig is available.
	KubeconfigPath(ctx context.Context) (string, error)

	// Name returns the logical cluster name.
	Name() string

	// Provider returns a short provider identifier (for example "kind" or "k3d").
	Provider() string
}

// Basic compile-time helper: ensure packages can refer to a nil Provider.
var _ Provider = (*noopProvider)(nil)

// noopProvider is an unexported no-op implementation used only to satisfy
// the var above; real providers should be implemented in provider-specific
// files (e.g. kind.go, k3d.go).
type noopProvider struct{}

func (*noopProvider) Create(ctx context.Context) error                   { return nil }
func (*noopProvider) Delete(ctx context.Context) error                   { return nil }
func (*noopProvider) Exists(ctx context.Context) (bool, error)           { return false, nil }
func (*noopProvider) KubeconfigPath(ctx context.Context) (string, error) { return "", nil }
func (*noopProvider) Name() string                                       { return "" }
func (*noopProvider) Provider() string                                   { return "noop" }

// internal interface implemented by providers in this package to accept dry-run.
type dryRunnable interface{ setDryRun(bool) }

// SetDryRun toggles dry-run on a provider if it supports it. No-op otherwise.
func SetDryRun(p Provider, dry bool) {
	if dr, ok := p.(dryRunnable); ok {
		dr.setDryRun(dry)
	}
}
