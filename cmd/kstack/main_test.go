package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// NOTE: These tests exercise Cobra command flows in dry-run mode to avoid
// external side effects. They primarily validate wiring and ensure codepaths
// execute without error.

func TestUp_DryRun_NoAddons(t *testing.T) {
	opts := &rootOptions{
		provider:    "kind",
		clusterName: "gc-test",
		addons:      "",
		namespace:   "gc",
		helmPath:    "helm",
		timeout:     10 * time.Second,
		dryRun:      true,
		noColor:     true,
	}
	cmd := newUpCmd(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd.SetContext(ctx)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("up dry-run (no addons) failed: %v", err)
	}
}

func TestUp_DryRun_WithExampleApp(t *testing.T) {
	opts := &rootOptions{
		provider:    "kind",
		clusterName: "gc-test",
		addons:      "example-app",
		namespace:   "app",
		helmPath:    "helm",
		timeout:     10 * time.Second,
		dryRun:      true,
		noColor:     true,
	}
	cmd := newUpCmd(opts)
	// set a context since newUpCmd references cmd.Context()
	cmd.SetContext(context.Background())
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("up dry-run (example-app) failed: %v", err)
	}
}

func TestDown_DryRun(t *testing.T) {
	opts := &rootOptions{
		provider:    "kind",
		clusterName: "gc-test",
		namespace:   "gc",
		timeout:     5 * time.Second,
		dryRun:      true,
		noColor:     true,
	}
	cmd := newDownCmd(opts)
	cmd.SetContext(context.Background())
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("down dry-run failed: %v", err)
	}
}

func TestAddons_Install_Uninstall_List_DryRun(t *testing.T) {
	opts := &rootOptions{
		helmPath: "helm",
		dryRun:   true,
		noColor:  true,
	}
	cmd := newAddonsCmd(opts)
	cmd.SetContext(context.Background())
	// install example-app
	cmd.SetArgs([]string{"install", "example-app"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("addons install dry-run failed: %v", err)
	}
	// list
	cmd = newAddonsCmd(opts)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("addons list failed: %v", err)
	}
	// uninstall example-app
	cmd = newAddonsCmd(opts)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"uninstall", "example-app"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("addons uninstall dry-run failed: %v", err)
	}
}

func TestPreflight_DryRun(t *testing.T) {
	opts := &rootOptions{helmPath: "helm", dryRun: true, noColor: true}
	cmd := newPreflightCmd(opts)
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("preflight dry-run failed: %v", err)
	}
}

func TestStatus_UnknownProvider_Error(t *testing.T) {
	opts := &rootOptions{provider: "does-not-exist", clusterName: "x", namespace: "y", helmPath: "helm", timeout: 1 * time.Second, dryRun: true}
	cmd := newStatusCmd(opts)
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error for unknown provider in status, got nil")
	}
}

func TestVersion_Prints(t *testing.T) {
	// Ensure version command runs without panic; we don't assert output text here
	version = "test-version"
	commit = "abcdef"
	buildDate = "2025-01-01T00:00:00Z"
	cmd := newVersionCmd()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("version execute failed: %v", err)
	}
}

func TestUp_DryRun_Postgres_HA(t *testing.T) {
	opts := &rootOptions{
		provider:    "kind",
		clusterName: "gc-test",
		addons:      "postgres",
		namespace:   "postgres",
		helmPath:    "helm",
		timeout:     10 * time.Second,
		dryRun:      true,
		noColor:     true,
	}
	cmd := newUpCmd(opts)
	cmd.Flags().Set("ha", "true")
	cmd.SetContext(context.Background())
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("up dry-run (postgres --ha) failed: %v", err)
	}
}

func TestUp_DryRun_InvalidSetPairs(t *testing.T) {
	opts := &rootOptions{provider: "kind", clusterName: "gc-test", addons: "example-app", namespace: "app", helmPath: "helm", timeout: 5 * time.Second, dryRun: true, noColor: true}
	cmd := newUpCmd(opts)
	cmd.SetContext(context.Background())
	// supply an invalid --set pair (no '=') which should cause an error
	cmd.Flags().Set("set", "not-a-pair")
	if err := cmd.RunE(cmd, nil); err == nil {
		t.Fatalf("expected error for invalid --set pair, got nil")
	}
}

func writeFake(t *testing.T, name, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("skipping fake scripts on windows")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(script), 0o755); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	// prepend to PATH so provider CLIs are found
	os.Setenv("PATH", dir+":"+os.Getenv("PATH"))
	return p
}

func TestStatus_WithFakeHelmAndProvider(t *testing.T) {
	// Fake helm emits a valid version and empty release list JSON
	helm := writeFake(t, "helm", "#!/usr/bin/env bash\nset -e\ncase \"$1\" in\n version) echo v3.12.1;;\n list) echo '[]';;\n *) echo ok;;\n esac\n")
	// Fake kind and docker to satisfy provider and preflight calls
	writeFake(t, "kind", "#!/usr/bin/env bash\nset -e\nif [ \"$1\" = \"get\" ] && [ \"$2\" = \"clusters\" ]; then echo none; exit 0; fi\nif [ \"$1\" = \"get\" ] && [ \"$2\" = \"kubeconfig\" ]; then echo apiVersion: v1; exit 0; fi\necho ok\n")
	writeFake(t, "docker", "#!/usr/bin/env bash\nexit 0\n")

	opts := &rootOptions{provider: "kind", clusterName: "gc", namespace: "gc", helmPath: helm, timeout: 2 * time.Second, dryRun: false, noColor: true}
	cmd := newStatusCmd(opts)
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status with fake helm/provider failed: %v", err)
	}
}

func TestAddons_Uninstall_WaitTimeout_DryRun(t *testing.T) {
	opts := &rootOptions{helmPath: "helm", dryRun: true}
	cmd := newAddonsCmd(opts)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"uninstall", "example-app", "--wait", "--helm-timeout", "45s"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("addons uninstall with wait/timeout failed: %v", err)
	}
}

func TestAddons_Install_InvalidValuesFile(t *testing.T) {
	opts := &rootOptions{helmPath: "helm", dryRun: true}
	cmd := newAddonsCmd(opts)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"install", "example-app", "--values", "/no/such/file.yaml"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error for missing values file")
	}
}

func TestDown_DryRun_PurgeAddons(t *testing.T) {
	opts := &rootOptions{provider: "kind", clusterName: "gc", namespace: "gc", helmPath: "helm", dryRun: true}
	cmd := newDownCmd(opts)
	cmd.SetContext(context.Background())
	cmd.Flags().Set("purge-addons", "true")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("down --purge-addons failed: %v", err)
	}
}

func TestAddons_Install_Prometheus_WithFlags_DryRun(t *testing.T) {
	opts := &rootOptions{helmPath: "helm", dryRun: true}
	cmd := newAddonsCmd(opts)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"install", "prometheus", "--wait", "--helm-timeout", "2m", "--atomic", "--set", "x=y"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("addons install prometheus with flags failed: %v", err)
	}
}

func TestStatus_WithFakeHelmList(t *testing.T) {
	helm := writeFake(t, "helm", "#!/usr/bin/env bash\nset -e\ncase \"$1\" in\n version) echo v3.12.1;;\n list) echo '[{\"name\":\"rel\",\"namespace\":\"gc\",\"revision\":\"1\",\"updated\":\"now\",\"status\":\"deployed\",\"chart\":\"x-1.0.0\",\"app_version\":\"1.0.0\"}]';;\n *) echo ok;;\n esac\n")
	writeFake(t, "kind", "#!/usr/bin/env bash\nset -e\nif [ \"$1\" = \"get\" ] && [ \"$2\" = \"clusters\" ]; then echo gc; fi\nif [ \"$1\" = \"get\" ] && [ \"$2\" = \"kubeconfig\" ]; then echo apiVersion: v1; fi\nexit 0\n")
	writeFake(t, "docker", "#!/usr/bin/env bash\nexit 0\n")
	opts := &rootOptions{provider: "kind", clusterName: "gc", namespace: "gc", helmPath: helm, timeout: 2 * time.Second}
	cmd := newStatusCmd(opts)
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status with helm list failed: %v", err)
	}
}

func TestPreflight_NonDryRun_WithFakes(t *testing.T) {
	helm := writeFake(t, "helm", "#!/usr/bin/env bash\nset -e\n[ \"$1\" = \"version\" ] && { echo v3.13.0; exit 0; }\necho ok\n")
	writeFake(t, "kind", "#!/usr/bin/env bash\nexit 0\n")
	writeFake(t, "docker", "#!/usr/bin/env bash\nif [ \"$1\" = \"info\" ]; then echo ok; exit 0; fi\nexit 0\n")
	opts := &rootOptions{provider: "kind", helmPath: helm, dryRun: false}
	cmd := newPreflightCmd(opts)
	cmd.SetContext(context.Background())
	// enable verbose and debug flags to exercise those code paths
	cmd.Flags().Set("verbose", "true")
	cmd.Flags().Set("debug", "true")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("preflight with fakes failed: %v", err)
	}
}

func TestStatus_HelmUnavailable_DoesNotFail(t *testing.T) {
	// Fake helm that fails version to trigger not-available branch
	helm := writeFake(t, "helm", "#!/usr/bin/env bash\nexit 1\n")
	writeFake(t, "kind", "#!/usr/bin/env bash\nset -e\nif [ \"$1\" = \"get\" ] && [ \"$2\" = \"clusters\" ]; then echo gc; exit 0; fi\nexit 0\n")
	writeFake(t, "docker", "#!/usr/bin/env bash\nexit 0\n")
	opts := &rootOptions{provider: "kind", clusterName: "gc", namespace: "gc", helmPath: helm, timeout: 2 * time.Second}
	cmd := newStatusCmd(opts)
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status should not fail when helm unavailable: %v", err)
	}
}

func TestAddons_Unknown_Addon_Errors(t *testing.T) {
	opts := &rootOptions{helmPath: "helm", dryRun: true}
	// install unknown
	cmd := newAddonsCmd(opts)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"install", "does-not-exist"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error for unknown addon install")
	}
	// uninstall unknown
	cmd = newAddonsCmd(opts)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"uninstall", "does-not-exist"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error for unknown addon uninstall")
	}
}

func TestUp_DryRun_InvalidValuesFile(t *testing.T) {
	opts := &rootOptions{provider: "kind", clusterName: "gc", addons: "example-app", namespace: "app", helmPath: "helm", dryRun: true}
	cmd := newUpCmd(opts)
	cmd.SetContext(context.Background())
	cmd.Flags().Set("values", "/no/such/file.yaml")
	if err := cmd.RunE(cmd, nil); err == nil {
		t.Fatalf("expected error for invalid values file in up")
	}
}

func TestUp_NonDryRun_ClusterExists_NoAddons(t *testing.T) {
	// Fake kind and docker; ensure Exists true and kubeconfig path succeeds
	writeFake(t, "kind", "#!/usr/bin/env bash\nset -e\nif [ \"$1\" = \"get\" ] && [ \"$2\" = \"clusters\" ]; then echo gc-exists; exit 0; fi\nif [ \"$1\" = \"get\" ] && [ \"$2\" = \"kubeconfig\" ]; then echo apiVersion: v1; exit 0; fi\nexit 0\n")
	writeFake(t, "docker", "#!/usr/bin/env bash\nif [ \"$1\" = \"info\" ]; then exit 0; fi\nexit 0\n")
	opts := &rootOptions{provider: "kind", clusterName: "gc-exists", addons: "", namespace: "gc", timeout: 2 * time.Second, dryRun: false}
	cmd := newUpCmd(opts)
	cmd.SetContext(context.Background())
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("up non-dry-run exists path failed: %v", err)
	}
}

func TestAddons_Install_LocalChart_NonDryRun(t *testing.T) {
	// Fake helm that supports version and upgrade; repo commands would fail if invoked
	helm := writeFake(t, "helm", "#!/usr/bin/env bash\nset -e\ncase \"$1\" in\n version) echo v3.12.1;;\n upgrade) echo ok;;\n *) exit 1;;\n esac\n")
	opts := &rootOptions{helmPath: helm, dryRun: false}
	cmd := newAddonsCmd(opts)
	cmd.SetContext(context.Background())
	// Install example-app which uses a local chart path, so no repo add/update should be attempted
	cmd.SetArgs([]string{"install", "example-app"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("addons install local chart failed: %v", err)
	}
}

func TestUp_DryRun_UnknownAddon_Error(t *testing.T) {
	opts := &rootOptions{provider: "kind", clusterName: "gc", addons: "does-not-exist", namespace: "gc", helmPath: "helm", dryRun: true}
	cmd := newUpCmd(opts)
	cmd.SetContext(context.Background())
	if err := cmd.RunE(cmd, nil); err == nil {
		t.Fatalf("expected error for unknown addon in up")
	}
}

func TestStatus_KubeconfigPrinted_WithFakes(t *testing.T) {
	helm := writeFake(t, "helm", "#!/usr/bin/env bash\nset -e\ncase \"$1\" in\n version) echo v3.12.1;;\n list) echo '[]';;\n *) echo ok;;\n esac\n")
	writeFake(t, "kind", "#!/usr/bin/env bash\nset -e\nif [ \"$1\" = \"get\" ] && [ \"$2\" = \"clusters\" ]; then echo gc-k; exit 0; fi\nif [ \"$1\" = \"get\" ] && [ \"$2\" = \"kubeconfig\" ]; then echo apiVersion: v1; exit 0; fi\nexit 0\n")
	opts := &rootOptions{provider: "kind", clusterName: "gc-k", namespace: "gc", helmPath: helm}
	cmd := newStatusCmd(opts)
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status kubeconfig path test failed: %v", err)
	}
}

func TestAddons_Install_Postgres_HA_DryRun(t *testing.T) {
	opts := &rootOptions{helmPath: "helm", dryRun: true}
	cmd := newAddonsCmd(opts)
	cmd.SetContext(context.Background())
	cmd.SetArgs([]string{"install", "postgres", "--ha"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("addons install postgres --ha failed: %v", err)
	}
}

func TestDown_NonDryRun_PurgeAddons_WithFakes(t *testing.T) {
	helm := writeFake(t, "helm", "#!/usr/bin/env bash\nset -e\ncase \"$1\" in\n version) echo v3.13.0;;\n uninstall) echo ok;;\n *) echo ok;;\n esac\n")
	writeFake(t, "kind", "#!/usr/bin/env bash\nset -e\nif [ \"$1\" = \"delete\" ] && [ \"$2\" = \"cluster\" ]; then exit 0; fi\necho ok\n")
	opts := &rootOptions{provider: "kind", clusterName: "gc", namespace: "gc", helmPath: helm, dryRun: false, timeout: 2 * time.Second}
	cmd := newDownCmd(opts)
	cmd.SetContext(context.Background())
	cmd.Flags().Set("purge-addons", "true")
	if err := cmd.Execute(); err != nil {
		t.Fatalf("down non-dry-run with fakes failed: %v", err)
	}
}

func TestStatus_HelmListParseError(t *testing.T) {
	helm := writeFake(t, "helm", "#!/usr/bin/env bash\nset -e\ncase \"$1\" in\n version) echo v3.12.1;;\n list) echo not-json;;\n *) echo ok;;\n esac\n")
	writeFake(t, "kind", "#!/usr/bin/env bash\nset -e\nif [ \"$1\" = \"get\" ] && [ \"$2\" = \"clusters\" ]; then echo gc; exit 0; fi\nexit 0\n")
	writeFake(t, "docker", "#!/usr/bin/env bash\nexit 0\n")
	opts := &rootOptions{provider: "kind", clusterName: "gc", namespace: "gc", helmPath: helm}
	cmd := newStatusCmd(opts)
	cmd.SetContext(context.Background())
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status should not fail on helm list parse error: %v", err)
	}
}
