package addons

import (
	"fmt"
	"strings"
)

// Addon represents a Helm-installable addon.
type Addon interface {
	Name() string
	Chart() string
	RepoName() string
	RepoURL() string
	Namespace() string
	ValuesFiles() []string
}

var registry = map[string]Addon{}

// Register registers an addon in the global registry.
func Register(a Addon) {
	registry[a.Name()] = a
}

// Get returns a registered addon by name.
func Get(name string) (Addon, error) {
	if a, ok := registry[name]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("addon not found: %s", name)
}

// List returns names of registered addons.
func List() []string {
	out := make([]string, 0, len(registry))
	for k := range registry {
		out = append(out, k)
	}
	return out
}

// ParseList parses a comma-separated addons string into slice.
func ParseList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
