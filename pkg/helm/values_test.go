package helm

import (
	"os"
	"path/filepath"
	"testing"

	yaml "gopkg.in/yaml.v3"
)

// helper to read a YAML file into a generic map
func readYAMLFileToMap(t *testing.T, path string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed reading %s: %v", path, err)
	}
	var out map[string]any
	if err := yaml.Unmarshal(b, &out); err != nil {
		t.Fatalf("failed unmarshalling %s: %v", path, err)
	}
	if out == nil {
		out = map[string]any{}
	}
	return out
}

func TestMergeValues_EmptyInputs(t *testing.T) {
	mergedPath, cleanup, err := MergeValues(nil, nil)
	if err != nil {
		t.Fatalf("MergeValues returned error: %v", err)
	}
	if mergedPath == "" {
		t.Fatalf("expected a merged file path, got empty")
	}
	// merged YAML should be an empty map
	m := readYAMLFileToMap(t, mergedPath)
	if len(m) != 0 {
		t.Fatalf("expected empty map, got: %#v", m)
	}
	// cleanup should remove merged file
	if err := cleanup(); err != nil {
		t.Fatalf("cleanup returned error: %v", err)
	}
	if _, err := os.Stat(mergedPath); !os.IsNotExist(err) {
		t.Fatalf("expected merged file to be removed, stat err: %v", err)
	}
}

func TestMergeValues_MergesFilesAndOverrides(t *testing.T) {
	dir := t.TempDir()
	f1, err := os.CreateTemp(dir, "v1-*.yaml")
	if err != nil {
		t.Fatalf("create temp file1: %v", err)
	}
	defer f1.Close()
	f2, err := os.CreateTemp(dir, "v2-*.yaml")
	if err != nil {
		t.Fatalf("create temp file2: %v", err)
	}
	defer f2.Close()

	// file 1 values
	if _, err := f1.WriteString(
		"a: 1\n" +
			"b:\n  x: foo\n  y: bar\n" +
			"c:\n  - 1\n" +
			"d: old\n",
	); err != nil {
		t.Fatalf("write f1: %v", err)
	}

	// file 2 values override/extend maps, replace lists
	if _, err := f2.WriteString(
		"b:\n  y: baz\n  z: qux\n" +
			"c:\n  - 2\n" +
			"e:\n  f: 3\n",
	); err != nil {
		t.Fatalf("write f2: %v", err)
	}

	overrides := map[string]any{
		"a":   42,
		"b":   map[string]any{"z": "override"},
		"new": "yes",
	}

	mergedPath, cleanup, err := MergeValues([]string{f1.Name(), f2.Name()}, overrides)
	if err != nil {
		t.Fatalf("MergeValues returned error: %v", err)
	}
	defer func() { _ = cleanup() }()

	m := readYAMLFileToMap(t, mergedPath)

	// a should be overridden to 42
	if got, ok := m["a"].(int); !ok || got != 42 {
		t.Fatalf("expected a=42 (int), got: %#v", m["a"])
	}

	// b should be merged recursively: x: foo (from f1), y: baz (from f2), z: override (from overrides)
	bMap, ok := m["b"].(map[string]any)
	if !ok {
		t.Fatalf("expected b to be a map, got: %#v", m["b"])
	}
	if bMap["x"] != "foo" || bMap["y"] != "baz" || bMap["z"] != "override" {
		t.Fatalf("unexpected b map: %#v", bMap)
	}

	// c should be replaced list from f2: [2]
	cList, ok := m["c"].([]any)
	if !ok || len(cList) != 1 {
		t.Fatalf("expected c to be a single-item list, got: %#v", m["c"])
	}
	if v, ok := cList[0].(int); !ok || v != 2 {
		t.Fatalf("expected c[0]=2 (int), got: %#v", cList[0])
	}

	// d should persist from f1
	if m["d"] != "old" {
		t.Fatalf("expected d=old, got: %#v", m["d"])
	}

	// e.f should be 3
	eMap, ok := m["e"].(map[string]any)
	if !ok {
		t.Fatalf("expected e to be a map, got: %#v", m["e"])
	}
	if f, ok := eMap["f"].(int); !ok || f != 3 {
		t.Fatalf("expected e.f=3 (int), got: %#v", eMap["f"])
	}

	// new should be present from overrides
	if m["new"] != "yes" {
		t.Fatalf("expected new=yes, got: %#v", m["new"])
	}
}

func TestMergeValues_CleanupRemovesTempInputsOnly(t *testing.T) {
	// temp file under OS temp dir (will be removed)
	tmpFile, err := os.CreateTemp("", "mv-in-*.yaml")
	if err != nil {
		t.Fatalf("create temp input: %v", err)
	}
	tmpPath := tmpFile.Name()
	_, _ = tmpFile.WriteString("k: v\n")
	_ = tmpFile.Close()

	// non-temp file under current working directory (should not be removed)
	nonTempPath := filepath.Join(".", "mv-non-temp.yaml")
	if err := os.WriteFile(nonTempPath, []byte("x: y\n"), 0o644); err != nil {
		t.Fatalf("write non-temp input: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(nonTempPath) })

	mergedPath, cleanup, err := MergeValues([]string{tmpPath, nonTempPath}, nil)
	if err != nil {
		t.Fatalf("MergeValues returned error: %v", err)
	}

	// run cleanup
	if err := cleanup(); err != nil {
		t.Fatalf("cleanup returned error: %v", err)
	}

	// merged file should be gone
	if _, err := os.Stat(mergedPath); !os.IsNotExist(err) {
		t.Fatalf("expected merged file removed, stat err: %v", err)
	}
	// temp input should be gone
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Fatalf("expected temp input removed, stat err: %v", err)
	}
	// non-temp input should still exist
	if _, err := os.Stat(nonTempPath); err != nil {
		t.Fatalf("expected non-temp input to exist, stat err: %v", err)
	}
}
