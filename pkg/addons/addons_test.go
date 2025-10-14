package addons_test

import (
	"reflect"
	"testing"

	addons "github.com/christk1/kstack/pkg/addons"
	_ "github.com/christk1/kstack/pkg/addons/exampleapp"
	_ "github.com/christk1/kstack/pkg/addons/grafana"
	_ "github.com/christk1/kstack/pkg/addons/kafka"
	_ "github.com/christk1/kstack/pkg/addons/postgres"
	_ "github.com/christk1/kstack/pkg/addons/prometheus"
)

func TestParseList(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want []string
	}{
		{name: "empty", in: "", want: nil},
		{name: "spaces only", in: "   \t  ", want: nil},
		{name: "single", in: "kafka", want: []string{"kafka"}},
		{name: "trim spaces", in: " kafka , kafka2 ", want: []string{"kafka", "kafka2"}},
		{name: "dedupe not required", in: "kafka,kafka", want: []string{"kafka", "kafka"}},
		{name: "ignore empty entries", in: "kafka,,kafka2, ", want: []string{"kafka", "kafka2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addons.ParseList(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("ParseList(%q) = %#v, want %#v", tt.in, got, tt.want)
			}
		})
	}
}

func TestRegistry_Basics(t *testing.T) {
	// Built-in addons should be registered via init()
	names := addons.List()
	// Expect at least these
	want := map[string]bool{"prometheus": true, "kafka": true, "postgres": true, "example-app": true, "grafana": true}
	for w := range want {
		found := false
		for _, n := range names {
			if n == w {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected addon %q to be registered; got %v", w, names)
		}
		if _, err := addons.Get(w); err != nil {
			t.Fatalf("Get(%q) failed: %v", w, err)
		}
	}
	if _, err := addons.Get("does-not-exist"); err == nil {
		t.Fatalf("expected error for unknown addon")
	}
}
