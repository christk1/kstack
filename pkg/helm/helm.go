package helm

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/christk1/kstack/utils"
)

// HelmClient is a thin wrapper around the helm binary. It shells out to
// `helm` and exposes a minimal API used by the addons flow.
type HelmClient struct {
	Path   string
	DryRun bool
}

// NewClient returns a HelmClient using the provided helm binary path.
func NewClient(path string) *HelmClient { return &HelmClient{Path: path} }

// Preflight checks that the helm binary is available and returns its short
// version string (e.g. "v3.12.0+g...") or an error.
func (h *HelmClient) Preflight(timeout time.Duration) (string, error) {
	if h.DryRun {
		utils.Info("DRY-RUN: %s version --short", h.Path)
		return "dry-run", nil
	}
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, h.Path, "version", "--short")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("helm preflight failed: %w: %s", err, string(out))
	}
	return string(out), nil
}

// RepoAdd adds a helm repo: `helm repo add name url`.
func (h *HelmClient) RepoAdd(name, url string) error {
	if h.DryRun {
		utils.Info("DRY-RUN: %s repo add %s %s", h.Path, name, url)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, h.Path, "repo", "add", name, url)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("helm repo add failed: %w: %s", err, string(out))
	}
	return nil
}

// RepoUpdate runs `helm repo update`.
func (h *HelmClient) RepoUpdate() error {
	if h.DryRun {
		utils.Info("DRY-RUN: %s repo update", h.Path)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, h.Path, "repo", "update")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("helm repo update failed: %w: %s", err, string(out))
	}
	return nil
}

// InstallOrUpgrade runs `helm upgrade --install`.
// If wait is true, it adds `--wait` and `--timeout` with the provided timeout duration.
// If atomic is true, it adds `--atomic`. setPairs is a list of `key=val` pairs
// passed to `--set` and can be empty.
func (h *HelmClient) InstallOrUpgrade(release, chart, namespace, valuesFile string, wait bool, timeout time.Duration, atomic bool, setPairs []string) error {
	args := []string{"upgrade", "--install", release, chart, "-n", namespace}
	if valuesFile != "" {
		args = append(args, "-f", valuesFile)
	}
	args = append(args, "--create-namespace")
	if atomic {
		args = append(args, "--atomic")
	}
	for _, s := range setPairs {
		if s != "" {
			args = append(args, "--set", s)
		}
	}
	if wait {
		args = append(args, "--wait", "--timeout", timeout.String())
	}

	if h.DryRun {
		utils.Info("DRY-RUN: %s %s", h.Path, strings.Join(args, " "))
		return nil
	}

	// Use provided timeout + a small buffer for the command context
	ctxTimeout := timeout
	if ctxTimeout < 30*time.Second {
		ctxTimeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, h.Path, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("helm upgrade --install failed: %w: %s", err, string(out))
	}
	return nil
}

// Uninstall runs `helm uninstall <release> -n <ns>`.
// If wait is true, it includes `--timeout`.
func (h *HelmClient) Uninstall(release, namespace string, wait bool, timeout time.Duration) error {
	args := []string{"uninstall", release, "-n", namespace}
	if wait {
		args = append(args, "--timeout", timeout.String())
	}
	if h.DryRun {
		utils.Info("DRY-RUN: %s %s", h.Path, strings.Join(args, " "))
		return nil
	}
	ctxTimeout := timeout
	if ctxTimeout < 30*time.Second {
		ctxTimeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), ctxTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, h.Path, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("helm uninstall failed: %w: %s", err, string(out))
	}
	return nil
}

// ReleaseInfo is a minimal representation of a Helm release as returned by
// `helm list -n <ns> -o json`.
type ReleaseInfo struct {
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	Revision   string `json:"revision"`
	Updated    string `json:"updated"`
	Status     string `json:"status"`
	Chart      string `json:"chart"`
	AppVersion string `json:"app_version"`
}

// ListReleases returns Helm releases in the given namespace by calling
// `helm list -n <ns> -o json` and parsing the JSON output.
func (h *HelmClient) ListReleases(namespace string) ([]ReleaseInfo, error) {
	if h.DryRun {
		utils.Info("DRY-RUN: %s list -n %s -o json", h.Path, namespace)
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	args := []string{"list", "-n", namespace, "-o", "json"}
	cmd := exec.CommandContext(ctx, h.Path, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("helm list failed: %w: %s", err, string(out))
	}
	var releases []ReleaseInfo
	if err := json.Unmarshal(out, &releases); err != nil {
		return nil, fmt.Errorf("parse helm list output: %w: %s", err, string(out))
	}
	return releases, nil
}
