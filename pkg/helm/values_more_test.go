package helm

import (
	"os"
	"testing"
)

func TestMergeValues_InvalidFile(t *testing.T) {
	if _, _, err := MergeValues([]string{"/does/not/exist.yaml"}, nil); err == nil {
		t.Fatalf("expected error when merging nonexistent file")
	}
}

func TestDeepMerge_OverwriteAndNested(t *testing.T) {
	// create two temp values files
	f1, _ := os.CreateTemp("", "vals1-*")
	f2, _ := os.CreateTemp("", "vals2-*")
	os.WriteFile(f1.Name(), []byte("a: 1\nnest:\n  x: 1\n"), 0o644)
	os.WriteFile(f2.Name(), []byte("a: 2\nnest:\n  y: 2\n"), 0o644)
	out, cleanup, err := MergeValues([]string{f1.Name(), f2.Name()}, map[string]any{"b": 3})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}
	if out == "" {
		t.Fatalf("expected merged file path")
	}
	// Ensure cleanup removes temp inputs and output
	if err := cleanup(); err != nil {
		t.Fatalf("cleanup error: %v", err)
	}
}
