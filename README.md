<p align="center">
  <img src="assets/kubernetes-logo.png" alt="Kubernetes" height="80" />
  &nbsp;&nbsp;&nbsp;
  <img src="assets/go-gopher.png" alt="Go Gopher" height="80" />
  
</p>

<p align="center" style="font-size: 12px;">
  Go Gopher by Renee French (CC BY 3.0). Kubernetes is a registered trademark of The Linux Foundation®.
  Logos used here for identification purposes only. See `assets/ATTRIBUTION.md`.
</p>

## kstack

Developer-first CLI and reference template for spinning up a local Kubernetes stack. It creates kind or k3d clusters and installs Helm-based addons (Prometheus, Grafana, Kafka, Postgres) plus a small Example App.

Use it as-is for local development and demos, or as a starting point to tailor your own stack—the built-in addons are examples you can extend or remove. The project is cleanly layered (providers → Helm wrapper → addon registry) and optimized for fast, reproducible workflows.

<p align="center">
  <a href="https://github.com/christk1/kstack/actions/workflows/ci.yml">
    <img alt="CI" src="https://img.shields.io/github/actions/workflow/status/christk1/kstack/ci.yml?branch=master" />
  </a>
  <a href="https://github.com/christk1/kstack/releases">
    <img alt="Release" src="https://img.shields.io/github/v/release/christk1/kstack?sort=semver" />
  </a>
  <a href="go.mod">
    <img alt="Go version" src="https://img.shields.io/github/go-mod/go-version/christk1/kstack" />
  </a>
  <a href="LICENSE">
    <img alt="License" src="https://img.shields.io/github/license/christk1/kstack" />
  </a>
  <a href="https://codecov.io/gh/christk1/kstack">
    <img alt="Coverage" src="https://codecov.io/gh/christk1/kstack/graph/badge.svg" />
  </a>
</p>

---

## Requirements

- Go 1.25 (toolchain) — for building and `go install`
- Docker — installed and daemon running
- Helm 3 — on PATH
- One provider CLI: kind or k3d — on PATH

Optional: kubectl (to port-forward and inspect resources)

---

## Install

Install the CLI in your GOPATH/bin (requires Go 1.25):

```bash
go install github.com/christk1/kstack/cmd/kstack@latest
```

Or build from source:

```bash
git clone https://github.com/christk1/kstack.git
cd kstack
go mod tidy
go build -o kstack ./cmd/kstack
```

Verify environment:

```bash
./kstack preflight --verbose
```

---

## Quick start (linear)

1) Create a cluster and install core addons (Prometheus + Grafana)

```bash
./kstack up --addons prometheus,grafana
```

2) Install the Example App (requires a local image)

The chart defaults to image `my-app:latest`. The recommended quick path is to use the Makefile which builds, loads into kind, and installs the chart:

```bash
# recommended: build, load into kind, and install in one step
make demo-deploy

# You can override the demo image used by the Makefile. DEMO_IMAGE defaults to `my-app:latest`.
# Example: make DEMO_IMAGE=example-app:local demo-deploy
```

If you prefer to run the steps manually, build and load the image for your provider below.

- kind (default cluster name: kstack; change if needed):

```bash
docker build -t my-app:latest ./pkg/addons/exampleapp/demo
kind load docker-image my-app:latest --name kstack
./kstack addons install example-app --wait
```

- k3d (replace cluster name if different):

```bash
docker build -t my-app:latest ./pkg/addons/exampleapp/demo
k3d image import my-app:latest -c kstack
./kstack addons install example-app --wait
```

3) Port-forward to explore (helper script)

Start background port-forwards for Grafana, Prometheus, and the Example App using the helper script:

```bash
bash scripts/port-forward.sh start
```

Stop them later with:

```bash
bash scripts/port-forward.sh stop
```

- Check current status (PID files):

```bash
bash scripts/port-forward.sh status
```

- Logs: /tmp/kstack-portforward-logs/*.log
- PIDs: /tmp/kstack-portforward-logs/*.pid

URLs

- Grafana: http://localhost:3000
- Prometheus: http://localhost:9090

Grafana credentials

The Grafana chart creates an admin user whose password is stored in a Kubernetes secret. To retrieve the password run:

```bash
# find the grafana secret and print the admin password
kubectl -n monitoring get secret --selector=app.kubernetes.io/name=grafana -o jsonpath='{.items[0].data.admin-password}' | base64 --decode; echo
```

The default username is `admin`.


## Testing the Example App

Generate simple traffic against the example app to verify it’s reachable and to populate metrics.

1) Ensure the port-forwards are running (see step above), then run the provided script to hit the app every second:

```bash
bash scripts/hit-example-app.sh http://localhost:8080 1
```

- First argument is the URL (defaults to http://localhost:8080 if omitted)
- Second argument is the interval in seconds (defaults to 1)
- Stop with Ctrl-C

You can then explore metrics in Prometheus at http://localhost:9090.

Note about Grafana timing

Grafana panels display the data that Prometheus has already scraped and stored. It can take a minute or two for graphs to appear because:

- Prometheus scrapes targets on a schedule (scrape interval, commonly 15s or 30s).
- Many dashboard panels use functions like rate(...[1m]) which need several samples to compute a value.
- Grafana dashboard time range and auto-refresh affect visibility — set the range to the last 5 minutes and enable refresh while testing.

If graphs are empty right away, try the quick checks in Troubleshooting (query Prometheus or inspect targets) — usually data appears once Prometheus completes a few scrapes.


4) Check status

```bash
./kstack status
```

5) Uninstall and tear down

```bash
./kstack addons uninstall prometheus --wait --helm-timeout 5m
./kstack down --purge-addons
```

Tip: You can dry-run any command to preview actions:

```bash
./kstack up --addons kafka --dry-run
```

---

## Commands (overview)

- up — create cluster and install requested addons
- down — delete cluster (use `--purge-addons` to uninstall built-ins first)
- addons install|uninstall|list — manage individual addons
- status — show cluster existence and Helm releases in the namespace
- preflight — validate Docker, provider CLI, and Helm availability
- version — print build-time version metadata

Global flags:

- `--provider kind|k3d` (default: kind)
- `--cluster <name>` (default: kstack)
- `--addons <csv>` (e.g., `prometheus,postgres`)
- `--namespace <ns>` (default: kstack)
- `--kubeconfig <path>` (optional)
- `--helm <path>` (default: helm)
- `--timeout <dur>` (default: 10m)
- `--dry-run` — print planned actions only
- `-v, --verbose` / `--debug` — increase diagnostic output
- `--no-color` — disable ANSI colors

Addon install flags:

- `--values <file>` (repeatable)
- `--set key=val` (repeatable)
- `--wait` / `--helm-timeout <dur>`
- `--atomic`
- `--ha` (where supported; e.g., `postgres`)

Examples:

```bash
# Postgres (single-node)
./kstack up --addons postgres

# Postgres HA (bitnami/postgresql-ha)
./kstack up --addons postgres --ha

# Install Prometheus with atomic/wait
./kstack addons install prometheus --atomic --wait --helm-timeout 15m
```

---

## Built-in addons

- Prometheus (simple server chart)
  - Chart: `prometheus-community/prometheus`
  - Repo: `https://prometheus-community.github.io/helm-charts`
  - Namespace: `monitoring`

- Kafka (Bitnami)
  - Chart: `bitnami/kafka`
  - Repo: `https://charts.bitnami.com/bitnami`
  - Namespace: `kafka`

- Postgres (Bitnami)
  - Chart: `bitnami/postgresql` (or `bitnami/postgresql-ha` with `--ha`)
  - Repo: `https://charts.bitnami.com/bitnami`
  - Namespace: `postgres`

- Grafana
  - Chart: `grafana/grafana`
  - Repo: `https://grafana.github.io/helm-charts`
  - Namespace: `monitoring`

- Example App (local chart)
  - Chart: `./pkg/addons/exampleapp/chart`
  - Namespace: `app`
  - No repo add/update is required (local path)

Defaults are development-friendly (e.g., persistence disabled). Override with `--values` and `--set`.

Note on defaults: these built-ins are examples

The included addons are meant as practical examples for local stacks. You can:

- Extend them by editing their `values.yaml` under `pkg/addons/<name>/` or by passing overrides via `--values` and `--set`.
- Remove them from your own fork by deleting their folders under `pkg/addons/<name>/` and removing the corresponding blank imports in `cmd/kstack/main.go` (the `_ ".../pkg/addons/<name>"` lines that auto-register addons).
- Simply not install an addon by omitting it from `--addons` or by not calling `kstack addons install <name>`.

Add your own addon (quick recipe)

1) Create a folder: `pkg/addons/mychart/` and add a minimal `values.yaml` tailored for dev (persistence off, small resources).
2) Implement the Addon in `pkg/addons/mychart/mychart.go` and register it:

```go
package mychart

import "github.com/christk1/kstack/pkg/addons"

type addon struct{}

func (a *addon) Name() string      { return "mychart" }
func (a *addon) Chart() string     { return "bitnami/redis" } // or a local path: "./pkg/addons/mychart/chart"
func (a *addon) RepoName() string  { return "bitnami" }       // empty if Chart() is a local path
func (a *addon) RepoURL() string   { return "https://charts.bitnami.com/bitnami" }
func (a *addon) Namespace() string { return "redis" }
func (a *addon) ValuesFiles() []string { return []string{"pkg/addons/mychart/values.yaml"} }

func init() { addons.Register(&addon{}) }
```

3) If using a local chart, put it under `pkg/addons/mychart/chart/` with `Chart.yaml`, `templates/`, and set `Chart()` to that relative path. Leave `RepoName()/RepoURL()` empty so repo add/update is skipped.
4) Wire it into the CLI by adding a blank import in `cmd/kstack/main.go` alongside the others:

```go
_ "github.com/christk1/kstack/pkg/addons/mychart"
```

5) Install it:

```bash
./kstack addons install mychart --wait \
  --values pkg/addons/mychart/values.yaml \
  --set some.key=value
```

---

## Troubleshooting

- "docker not found": install Docker or ensure it is on PATH.
- "permission denied" from Docker: add your user to the `docker` group or run with appropriate privileges.
- "cannot connect to the Docker daemon": ensure Docker is running (e.g., `systemctl start docker`).
- "kind/k3d not found": install the chosen provider CLI and re-run `preflight`.
- "helm preflight failed": install Helm 3 and ensure it’s on PATH.

Use `kstack preflight --verbose` (or `--debug`) to print full outputs for CI and debugging.

---

## Development

Build and test (requires Go 1.25):

```bash
go build -o kstack ./cmd/kstack
go test ./...
```

Run with debug output:

```bash
./kstack up --addons prometheus -v --debug
```

Version metadata (example):

```bash
go build -ldflags "-X main.version=0.1.0 -X main.commit=$(git rev-parse --short HEAD) -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o kstack ./cmd/kstack
./kstack version
```

---

## License

Apache-2.0
