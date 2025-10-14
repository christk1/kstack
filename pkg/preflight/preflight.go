package preflight

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Verbose and Debug control whether preflight helpers print full command
// output. These can be toggled by callers (CLI) to surface extra info for CI
// or debugging scenarios.
var Verbose bool
var Debug bool

// CheckCommands ensures each executable name is present in PATH. Returns a
// friendly error listing missing commands.
func CheckCommands(cmds []string) error {
	var missing []string
	for _, c := range cmds {
		if c == "" {
			continue
		}
		if _, err := exec.LookPath(c); err != nil {
			missing = append(missing, c)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("required CLI tools not found in PATH: %s; please install them and try again", strings.Join(missing, ", "))
}

// CheckDocker runs `docker info` to ensure the Docker daemon is accessible.
// It uses the provided context for timeout/cancellation.
func CheckDocker(ctx context.Context) error {
	// ensure docker binary exists first
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found in PATH. Install Docker: https://docs.docker.com/get-docker/")
	}

	// run `docker info`
	// create a short-lived context if parent has none
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cctx, "docker", "info")
	out, err := cmd.CombinedOutput()
	if err != nil {
		// Provide actionable advice
		msg := strings.TrimSpace(string(out))
		advice := ""
		if strings.Contains(strings.ToLower(msg), "permission denied") {
			advice = " (permission denied — try adding your user to the docker group or run as root)"
		} else if strings.Contains(strings.ToLower(msg), "cannot connect") || strings.Contains(strings.ToLower(msg), "connect") {
			advice = " (is the Docker daemon running? try 'systemctl start docker' or ensure your environment is configured)"
		}
		// When verbose/debug is enabled, surface the raw output as well.
		if Verbose || Debug {
			return fmt.Errorf("docker is installed but not usable: %v: %s%s\n--- docker output ---\n%s", err, msg, advice, string(out))
		}
		return fmt.Errorf("docker is installed but not usable: %v: %s%s", err, msg, advice)
	}

	// On success, optionally print the full `docker info` output for debugging/CI.
	if Verbose || Debug {
		fmt.Printf("docker info output:\n%s\n", string(out))
	}
	return nil
}

// CheckProviderCLI validates provider-specific CLIs (kind or k3d).
func CheckProviderCLI(provider string) error {
	switch provider {
	case "kind":
		if _, err := exec.LookPath("kind"); err != nil {
			return fmt.Errorf("'kind' not found in PATH. Install kind: https://kind.sigs.k8s.io/")
		}
	case "k3d":
		if _, err := exec.LookPath("k3d"); err != nil {
			return fmt.Errorf("'k3d' not found in PATH. Install k3d: https://k3d.io/")
		}
	default:
		return fmt.Errorf("unknown provider '%s'", provider)
	}
	return nil
}
