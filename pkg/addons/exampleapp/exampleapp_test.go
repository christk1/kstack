package exampleapp

import "testing"

func TestExampleAppAddon_Basics(t *testing.T) {
	a := &exampleAppAddon{}
	if a.Name() != "example-app" || a.Chart() == "" || a.Namespace() != "app" {
		t.Fatalf("invalid fields: name=%s chart=%s ns=%s", a.Name(), a.Chart(), a.Namespace())
	}
	_ = a.ValuesFiles() // call for coverage
}
