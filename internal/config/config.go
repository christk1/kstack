package config

import (
	"os"
	"strings"
	"time"
)

// Config holds runtime options for kstack operations.
type Config struct {
	Provider    string
	ClusterName string
	Addons      []string
	Namespace   string
	Kubeconfig  string
	HelmPath    string
	Timeout     time.Duration
	Verbose     bool
	Debug       bool
}

// Defaults returns baseline defaults for the CLI.
func Defaults() Config {
	return Config{
		Provider:    "kind",
		ClusterName: "kstack",
		Namespace:   "kstack",
		HelmPath:    "helm",
		Timeout:     10 * time.Minute,
		Verbose:     false,
		Debug:       false,
	}
}

// FromEnv overlays values from environment variables (if set).
// Environment variables:
//
//	GO_CLOUD_PROVIDER, GO_CLOUD_CLUSTER, GO_CLOUD_ADDONS, GO_CLOUD_NAMESPACE,
//	GO_CLOUD_KUBECONFIG, GO_CLOUD_HELM, GO_CLOUD_TIMEOUT, GO_CLOUD_VERBOSE, GO_CLOUD_DEBUG
func FromEnv(base Config) Config {
	if v := os.Getenv("GO_CLOUD_PROVIDER"); v != "" {
		base.Provider = v
	}
	if v := os.Getenv("GO_CLOUD_CLUSTER"); v != "" {
		base.ClusterName = v
	}
	if v := os.Getenv("GO_CLOUD_ADDONS"); v != "" {
		base.Addons = ParseAddonsCSV(v)
	}
	if v := os.Getenv("GO_CLOUD_NAMESPACE"); v != "" {
		base.Namespace = v
	}
	if v := os.Getenv("GO_CLOUD_KUBECONFIG"); v != "" {
		base.Kubeconfig = v
	}
	if v := os.Getenv("GO_CLOUD_HELM"); v != "" {
		base.HelmPath = v
	}
	if v := os.Getenv("GO_CLOUD_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			base.Timeout = d
		}
	}
	if v := os.Getenv("GO_CLOUD_VERBOSE"); v != "" {
		base.Verbose = strings.EqualFold(v, "true") || v == "1"
	}
	if v := os.Getenv("GO_CLOUD_DEBUG"); v != "" {
		base.Debug = strings.EqualFold(v, "true") || v == "1"
	}
	return base
}

// ParseAddonsCSV converts a comma-separated string into a slice, trimming blanks.
func ParseAddonsCSV(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
