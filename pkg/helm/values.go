package helm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

// MergeValues reads the provided YAML files in order and deep-merges their
// contents. Optional overrides are merged last. The merged YAML is written
// to a temp file and the path is returned, along with a cleanup func that
// removes the merged file and any input files that live inside os.TempDir().
// The cleanup func returns an error if any removals fail.
func MergeValues(files []string, overrides map[string]any) (string, func() error, error) {
	merged := make(map[string]any)

	for _, f := range files {
		if f == "" {
			continue
		}
		b, err := os.ReadFile(f)
		if err != nil {
			return "", func() error { return nil }, fmt.Errorf("read values file %s: %w", f, err)
		}
		var m map[string]any
		if err := yaml.Unmarshal(b, &m); err != nil {
			return "", func() error { return nil }, fmt.Errorf("parse values file %s: %w", f, err)
		}
		deepMerge(merged, m)
	}

	if overrides != nil {
		deepMerge(merged, overrides)
	}

	out, err := yaml.Marshal(merged)
	if err != nil {
		return "", func() error { return nil }, fmt.Errorf("marshal merged values: %w", err)
	}

	tmpDir := os.TempDir()
	f, err := os.CreateTemp(tmpDir, "kstack-merged-values-")
	if err != nil {
		return "", func() error { return nil }, fmt.Errorf("create temp file: %w", err)
	}
	if _, err := f.Write(out); err != nil {
		f.Close()
		return "", func() error { return nil }, fmt.Errorf("write merged values: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", func() error { return nil }, fmt.Errorf("close merged values file: %w", err)
	}

	mergedPath := f.Name()

	cleanup := func() error {
		// remove merged file
		var cleanupErr error
		if mergedPath != "" {
			if err := os.Remove(mergedPath); err != nil {
				cleanupErr = err
			}
		}
		// remove input files if they are inside tmpDir
		for _, p := range files {
			if p == "" {
				continue
			}
			abs, err := filepath.Abs(p)
			if err != nil {
				continue
			}
			rel, err := filepath.Rel(tmpDir, abs)
			if err != nil {
				continue
			}
			if rel != "" && rel != "." && !strings.HasPrefix(rel, "..") {
				if err := os.Remove(abs); err != nil {
					cleanupErr = err
				}
			}
		}
		return cleanupErr
	}

	return mergedPath, cleanup, nil
}

// deepMerge merges src into dst recursively. Maps are merged; non-map values
// overwrite previous values.
func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if vMap, ok := v.(map[string]any); ok {
			if existing, found := dst[k]; found {
				if existingMap, ok2 := existing.(map[string]any); ok2 {
					deepMerge(existingMap, vMap)
					continue
				}
			}
			// copy map
			newMap := make(map[string]any)
			deepMerge(newMap, vMap)
			dst[k] = newMap
		} else {
			dst[k] = v
		}
	}
}
