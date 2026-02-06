# lgtm

Local implementation of the LGTM stack using Kind and Flux CD

## Stack Components

- **Mimir** - Metrics storage
- **Loki** - Log aggregation
- **Tempo** - Distributed tracing
- **Grafana** - Visualization and dashboards
- **Alloy** - Unified collector (metrics, logs, traces)
- **Alertmanager** - Alert management

## Bootstrap

To create the cluster and deploy the LGTM stack:

```bash
./bootstrap.sh
```

This will:
1. Create a 3-node Kind cluster
2. Install Flux CD
3. Configure Flux to pull from this repository via SSH
4. Deploy all LGTM components

## Access

After deployment, you can access Grafana:

```bash
kubectl port-forward svc/grafana 3000:3000 -n observability
```

Then open http://localhost:3000

Default credentials:
- Username: admin
- Password: Retrieve with: `kubectl get secret grafana -n observability -o jsonpath='{.data.admin-password}' | base64 -d`

## Structure

```
.
├── infrastructure/    # Infrastructure components (namespaces, repositories, Flux Kustomizations)
├── apps/             # LGTM stack applications (HelmReleases)
└── bootstrap.sh      # Bootstrap script
```

## Architecture

- **Infrastructure** Kustomization manages namespaces and Helm repositories
- **Apps** Kustomization deploys all LGTM components
- All components use filesystem storage (suitable for local development)
- Resource requests are minimized for Kind cluster compatibility

## Pre-configured Datasources

Grafana comes pre-configured with:
- Mimir (Prometheus-compatible metrics)
- Loki (logs)
- Tempo (traces)
- Alertmanager

## Pre-configured Dashboards

- Node Exporter Full
- Kubernetes Cluster
- Loki Dashboard
