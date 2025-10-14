package main

import (
	"io"
	"os"

	"github.com/christk1/kstack/pkg/addons"
	_ "github.com/christk1/kstack/pkg/addons/grafana"
)

func main() {
	// This small helper prints the generated values file from the grafana addon
	// to stdout, matching the behavior of ValuesFiles(). It is useful for CI
	// where we want a concrete values file to pass to `helm template`.
	a, err := addons.Get("grafana")
	if err != nil {
		panic(err)
	}
	files := a.ValuesFiles()
	if len(files) == 0 {
		panic("no values files returned")
	}
	f, err := os.Open(files[0])
	if err != nil {
		panic(err)
	}
	defer f.Close()
	io.Copy(os.Stdout, f)
}
