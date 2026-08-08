# Stackdome alpha observability

The API server exposes Prometheus metrics on `METRICS_BIND_ADDRESS` (default `0.0.0.0:9090`) at `/metrics`.

## Install

1. Add an internal-only port named `metrics` on the API pod and Service. Do not add this port to the public Ingress.
2. Configure Prometheus to scrape the port every 15 seconds with job name `stackdome-api-metrics`.
3. Import `grafana/alpha-overview.json` and select the Prometheus datasource when prompted.
4. Apply `prometheus-rules.yaml` in a namespace watched by Prometheus Operator.

To view the dashboard in Grafana:

1. Open **Dashboards → New → Import**.
2. Upload `grafana/alpha-overview.json`.
3. Select the Prometheus datasource that scrapes `stackdome-api-metrics`.
4. Select **Import**.

Example ServiceMonitor endpoint:

```yaml
endpoints:
  - port: metrics
    path: /metrics
    interval: 15s
```

## Local validation

Run:

```bash
make observability-check
```

The target installs the pinned Prometheus `v3.13.2` `promtool` binary at `bin/promtool`, validates the alert rules, and parses the Grafana dashboard JSON. The binary stays inside the ignored project `bin/` directory and is reused on later runs. Run `make promtool` when only the tool installation is needed.

## Five-minute smoke test

- Open `/metrics` through an internal port-forward and confirm `stackdome_http_server_requests_total`, `stackdome_stacks`, and `workqueue_depth` are present.
- Call `/health` and one API endpoint, then confirm the API row changes without a raw request ID or stack ID label.
- Trigger or wait for one worker execution and confirm its queue and outcome panels change.
- Temporarily point one test alert at a zero threshold and confirm it reaches the alpha alert destination, then restore the checked-in rule.
