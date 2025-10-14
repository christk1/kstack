# Grafana addon

This addon packages Grafana configuration and repo-managed dashboards.

Where dashboards live
- Repo-managed dashboards live in `pkg/addons/grafana/dashboards/` as JSON files.

How dashboards are provisioned
- At runtime the addon produces a temporary `values.yaml` that includes the contents of the `dashboards/` files under `dashboards.default.<name>.json` so the upstream `grafana/grafana` chart can render dashboard ConfigMaps from `values.dashboards`.
- The addon also enables the kiwigrid sidecar in values so a sidecar can pick up ConfigMaps labeled `grafana_dashboard` if you prefer sidecar-loading.

Contributor workflow
- Add a new dashboard JSON to `pkg/addons/grafana/dashboards/`.
- Update or run tests locally:

```bash
go test ./...   # ensures ValuesFiles injection works
```

- To manually verify Helm rendering locally:

```bash
go run ./scripts/generate-grafana-values | tee grafana-ci-values.yaml
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm template graf grafana/grafana -n monitoring -f grafana-ci-values.yaml --include-crds
```

Notes
- This repository currently injects dashboards into a temporary values file (keeps the addon small and easy to understand). If you prefer explicit ConfigMap templates per dashboard we can convert the addon into a wrapper chart which will render per-dashboard ConfigMaps from files in `dashboards/`.
