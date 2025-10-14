package cluster

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/christk1/kstack/utils"
)

// kindProvider implements Provider by invoking the `kind` CLI. It prefers
// to return actionable errors and uses context for timeouts.
type kindProvider struct {
	name   string
	DryRun bool
}

func NewKindProvider(name string) Provider { return &kindProvider{name: name} }

func (p *kindProvider) Create(ctx context.Context) error {
	// Run preflight checks first
	if p.DryRun {
		utils.Info("DRY-RUN: kind create cluster --name %s", p.name)
		return nil
	}
	if err := p.preflight(ctx); err != nil {
		return err
	}

	// kind create cluster --name <name>
	utils.Debug("running: kind create cluster --name %s", p.name)
	cmd := exec.CommandContext(ctx, "kind", "create", "cluster", "--name", p.name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kind create failed: %w: %s", err, string(out))
	}
	return nil
}

func (p *kindProvider) Delete(ctx context.Context) error {
	if p.DryRun {
		utils.Info("DRY-RUN: kind delete cluster --name %s", p.name)
		return nil
	}
	cmd := exec.CommandContext(ctx, "kind", "delete", "cluster", "--name", p.name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kind delete failed: %w: %s", err, string(out))
	}
	return nil
}

func (p *kindProvider) Exists(ctx context.Context) (bool, error) {
	if p.DryRun {
		utils.Info("DRY-RUN: kind get clusters (assume not exists)")
		return false, nil
	}
	// kind get clusters
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kind", "get", "clusters")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("kind get clusters failed: %w: %s", err, string(out))
	}
	lines := strings.Fields(string(out))
	for _, l := range lines {
		if l == p.name {
			return true, nil
		}
	}
	return false, nil
}

func (p *kindProvider) KubeconfigPath(ctx context.Context) (string, error) {
	if p.DryRun {
		utils.Info("DRY-RUN: kind get kubeconfig --name %s", p.name)
		return "", nil
	}
	// Run `kind get kubeconfig --name <name>` and write to a temp file.
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kind", "get", "kubeconfig", "--name", p.name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("kind get kubeconfig failed: %w: %s", err, string(out))
	}

	// Write to temp file
	tmp, err := os.CreateTemp("", "kind-kubeconfig-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp kubeconfig: %w", err)
	}
	defer tmp.Close()

	if _, err := tmp.Write(out); err != nil {
		return "", fmt.Errorf("failed to write kubeconfig to temp file: %w", err)
	}

	utils.Debug("wrote kind kubeconfig to %s", tmp.Name())
	return tmp.Name(), nil
}

// preflight verifies that required binaries and runtime (docker) are available.
func (p *kindProvider) preflight(ctx context.Context) error {
	// check kind binary
	ctx1, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx1, "kind", "--version").Run(); err != nil {
		return fmt.Errorf("kind CLI not found or not executable: %w", err)
	}

	// check docker availability
	ctx2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancel2()
	if err := exec.CommandContext(ctx2, "docker", "info").Run(); err != nil {
		return fmt.Errorf("docker does not appear to be available/running: %w", err)
	}

	return nil
}

func (p *kindProvider) Name() string     { return p.name }
func (p *kindProvider) Provider() string { return "kind" }

// setDryRun implements internal dryRunnable
func (p *kindProvider) setDryRun(d bool) { p.DryRun = d }
