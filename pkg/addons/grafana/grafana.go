package grafana

import (
	"embed"
	"os"
	"path"
	"strings"

	"github.com/christk1/kstack/pkg/addons"
)

//go:embed values.yaml
//go:embed dashboards/*
var fs embed.FS

type grafanaAddon struct{}

func (g *grafanaAddon) Name() string      { return "grafana" }
func (g *grafanaAddon) Chart() string     { return "grafana/grafana" }
func (g *grafanaAddon) RepoName() string  { return "grafana" }
func (g *grafanaAddon) RepoURL() string   { return "https://grafana.github.io/helm-charts" }
func (g *grafanaAddon) Namespace() string { return "monitoring" }
func (g *grafanaAddon) ValuesFiles() []string {
	// Read base values.yaml from embed FS
	b, err := fs.ReadFile("values.yaml")
	if err != nil {
		return nil
	}

	// Check for a dashboards file and inject it into a temp values file using
	// the helm chart's expected .Files.Get path by creating a small values
	// fragment that references the embedded file contents under the
	// dashboards.default.<name>.json.json key is not needed since we'll create
	// a templated ConfigMap in the chart using .Files.Get. To keep things
	// simple, append a small values fragment that the chart can ignore but
	// preserves the base values.

	// Create a temp dir to hold the values file and the dashboard file so that
	// helm can pick up .Files.Get when rendering from the filesystem.
	tmpDir, err := os.MkdirTemp("", "kstack-grafana-valuesdir-")
	if err != nil {
		return nil
	}

	// Prepare the base values content in memory (we'll write once after
	// optionally injecting dashboards). This avoids writing twice which can
	// produce duplicate keys if append logic runs incorrectly.
	basePath := path.Join(tmpDir, "values.yaml")
	baseContent := string(b)

	// Read embedded dashboards and inject them inline into the values.yaml so
	// the grafana chart will render them as ConfigMaps (the chart supports
	// creating dashboards from values.dashboards). We will either replace an
	// empty placeholder `dashboards: {}` in the base values file, or append a
	// dashboards.default.<name>.json block if no dashboards key exists. If the
	// base values already contain dashboards with content we will not inject to
	// avoid clobbering user-provided dashboards.
	entries, _ := fs.ReadDir("dashboards")
	if len(entries) > 0 {
		var dashYaml string
		dashYaml += "# Injected dashboards from repo\n"
		dashYaml += "dashboards:\n"
		dashYaml += "  default:\n"
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			content, err := fs.ReadFile(path.Join("dashboards", name))
			if err != nil {
				continue
			}
			// Create a YAML key name from the file name without extension
			key := name
			// strip extension if present
			if ext := path.Ext(name); ext != "" {
				key = name[:len(name)-len(ext)]
			}
			dashYaml += "    " + key + ":\n"
			dashYaml += "      json: |-\n"
			// indent each line of the JSON by 8 spaces
			lines := splitLines(string(content))
			for _, ln := range lines {
				dashYaml += "        " + ln + "\n"
			}
		}

		bStr := baseContent
		if strings.Contains(bStr, "dashboards: {}") {
			// Replace the empty placeholder with our dashboards block
			baseContent = strings.Replace(bStr, "dashboards: {}", dashYaml, 1)
		} else if strings.Contains(bStr, "dashboards:") {
			// dashboards exist with content; skip injection to avoid duplicates
			// write baseContent as-is
		} else {
			// Append dashboards YAML to the base values content
			baseContent = baseContent + "\n" + dashYaml
		}
	}

	// Finally write the computed baseContent to the temp file once.
	if err := os.WriteFile(basePath, []byte(baseContent), 0644); err != nil {
		return nil
	}

	// Return the path to the values file in the temp dir.
	return []string{basePath}
}

func init() {
	addons.Register(&grafanaAddon{})
}

// splitLines splits a string into lines, preserving empty lines.
func splitLines(s string) []string {
	var lines []string
	cur := ""
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '\n' {
			lines = append(lines, cur)
			cur = ""
			continue
		}
		cur += string(c)
	}
	// append last line
	lines = append(lines, cur)
	return lines
}
