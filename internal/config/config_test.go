package config

import (
	"reflect"
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	d := Defaults()
	if d.Provider != "kind" || d.ClusterName != "kstack" || d.Namespace != "kstack" || d.HelmPath != "helm" {
		t.Fatalf("unexpected defaults: %#v", d)
	}
	if d.Timeout <= 0 {
		t.Fatalf("expected positive default timeout, got %v", d.Timeout)
	}
}

func TestFromEnv_Overlay(t *testing.T) {
	t.Setenv("GO_CLOUD_PROVIDER", "k3d")
	t.Setenv("GO_CLOUD_CLUSTER", "dev")
	t.Setenv("GO_CLOUD_ADDONS", "kafka")
	t.Setenv("GO_CLOUD_NAMESPACE", "ns")
	t.Setenv("GO_CLOUD_KUBECONFIG", "/tmp/kubeconfig")
	t.Setenv("GO_CLOUD_HELM", "/usr/local/bin/helm")
	t.Setenv("GO_CLOUD_TIMEOUT", "45s")
	t.Setenv("GO_CLOUD_VERBOSE", "true")
	t.Setenv("GO_CLOUD_DEBUG", "1")

	got := FromEnv(Defaults())
	if got.Provider != "k3d" || got.ClusterName != "dev" || got.Namespace != "ns" || got.Kubeconfig != "/tmp/kubeconfig" || got.HelmPath != "/usr/local/bin/helm" {
		t.Fatalf("unexpected overlay: %#v", got)
	}
	if !reflect.DeepEqual(got.Addons, []string{"kafka"}) {
		t.Fatalf("unexpected addons: %#v", got.Addons)
	}
	if got.Timeout != 45*time.Second {
		t.Fatalf("unexpected timeout: %v", got.Timeout)
	}
	if !got.Verbose || !got.Debug {
		t.Fatalf("expected verbose and debug to be true: %#v", got)
	}
}

func TestParseAddonsCSV(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"kafka", []string{"kafka"}},
		{" kafka, kafka2 ", []string{"kafka", "kafka2"}},
		{"kafka,,kafka2,", []string{"kafka", "kafka2"}},
	}
	for _, c := range cases {
		got := ParseAddonsCSV(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Fatalf("ParseAddonsCSV(%q) = %#v, want %#v", c.in, got, c.want)
		}
	}
}
