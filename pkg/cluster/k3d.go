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

// k3dProvider implements Provider by invoking the `k3d` CLI.
type k3dProvider struct {
	name   string
	DryRun bool
}

func NewK3dProvider(name string) Provider { return &k3dProvider{name: name} }

func (p *k3dProvider) Create(ctx context.Context) error {
	if p.DryRun {
		utils.Info("DRY-RUN: k3d cluster create %s", p.name)
		return nil
	}
	if err := p.preflight(ctx); err != nil {
		return err
	}
	utils.Debug("running: k3d cluster create %s", p.name)
	cmd := exec.CommandContext(ctx, "k3d", "cluster", "create", p.name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("k3d create failed: %w: %s", err, string(out))
	}
	return nil
}

func (p *k3dProvider) Delete(ctx context.Context) error {
	if p.DryRun {
		utils.Info("DRY-RUN: k3d cluster delete %s", p.name)
		return nil
	}
	cmd := exec.CommandContext(ctx, "k3d", "cluster", "delete", p.name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("k3d delete failed: %w: %s", err, string(out))
	}
	return nil
}

func (p *k3dProvider) Exists(ctx context.Context) (bool, error) {
	if p.DryRun {
		utils.Info("DRY-RUN: k3d cluster list (assume not exists)")
		return false, nil
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "k3d", "cluster", "list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("k3d list failed: %w: %s", err, string(out))
	}
	lines := strings.Fields(string(out))
	for _, l := range lines {
		if l == p.name {
			return true, nil
		}
	}
	return false, nil
}

func (p *k3dProvider) KubeconfigPath(ctx context.Context) (string, error) {
	if p.DryRun {
		utils.Info("DRY-RUN: k3d kubeconfig get %s", p.name)
		return "", nil
	}
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "k3d", "kubeconfig", "get", p.name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("k3d get kubeconfig failed: %w: %s", err, string(out))
	}
	tmp, err := os.CreateTemp("", "k3d-kubeconfig-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp kubeconfig: %w", err)
	}
	defer tmp.Close()
	if _, err := tmp.Write(out); err != nil {
		return "", fmt.Errorf("failed to write kubeconfig to temp file: %w", err)
	}
	utils.Debug("wrote k3d kubeconfig to %s", tmp.Name())
	return tmp.Name(), nil
}

func (p *k3dProvider) preflight(ctx context.Context) error {
	ctx1, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx1, "k3d", "version").Run(); err != nil {
		return fmt.Errorf("k3d CLI not found or not executable: %w", err)
	}
	ctx2, cancel2 := context.WithTimeout(ctx, 5*time.Second)
	defer cancel2()
	if err := exec.CommandContext(ctx2, "docker", "info").Run(); err != nil {
		return fmt.Errorf("docker does not appear to be available/running: %w", err)
	}
	return nil
}

func (p *k3dProvider) Name() string     { return p.name }
func (p *k3dProvider) Provider() string { return "k3d" }

// setDryRun implements internal dryRunnable
func (p *k3dProvider) setDryRun(d bool) { p.DryRun = d }
