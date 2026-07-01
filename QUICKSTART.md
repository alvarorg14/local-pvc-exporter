# Quick Start

Get **local-pvc-exporter** running on your Kubernetes cluster in minutes.

## Prerequisites

- A Kubernetes cluster with `hostPath` or `local` PersistentVolumes
- [Helm 3](https://helm.sh/docs/intro/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/) configured for your cluster

## Install

Container images are published to `ghcr.io/alvarorg14/local-pvc-exporter`. The Helm chart is published to `oci://ghcr.io/alvarorg14/charts/local-pvc-exporter`.

Replace `X.Y.Z` with the [latest release](https://github.com/alvarorg14/local-pvc-exporter/releases) version:

```bash
helm install local-pvc-exporter oci://ghcr.io/alvarorg14/charts/local-pvc-exporter \
  --version X.Y.Z \
  --namespace monitoring \
  --create-namespace
```

Or install from the chart source in this repository:

```bash
helm install local-pvc-exporter ./charts/local-pvc-exporter \
  --namespace monitoring \
  --create-namespace
```

## Enable Prometheus Scraping

If you use the [Prometheus Operator](https://prometheus-operator.dev/), enable the ServiceMonitor:

```bash
helm install local-pvc-exporter oci://ghcr.io/alvarorg14/charts/local-pvc-exporter \
  --version X.Y.Z \
  --namespace monitoring \
  --create-namespace \
  --set serviceMonitor.enabled=true
```

## Verify

Confirm the DaemonSet pods are running:

```bash
kubectl get pods -n monitoring -l app.kubernetes.io/name=local-pvc-exporter
```

Port-forward to a pod and fetch metrics:

```bash
kubectl port-forward -n monitoring svc/local-pvc-exporter 8080:8080
curl -s localhost:8080/metrics | grep local_pvc
```

Check exporter logs for a successful scrape:

```bash
kubectl logs -n monitoring -l app.kubernetes.io/name=local-pvc-exporter --tail=20
```

Look for a log line like `scrape complete` with `errors: 0`.

## Try Some PromQL

Once Prometheus is scraping the exporter:

```promql
# PVC usage ratio
local_pvc_used_ratio{namespace="default"}

# PVCs over 80% full
local_pvc_used_ratio > 0.8

# Available space in bytes
local_pvc_available_bytes{persistentvolumeclaim="my-data"}
```

## Common Tweaks

Change the output unit or scrape interval at install time:

```bash
helm upgrade local-pvc-exporter oci://ghcr.io/alvarorg14/charts/local-pvc-exporter \
  --version X.Y.Z \
  --namespace monitoring \
  --reuse-values \
  --set unit=gib \
  --set scrapeInterval=2m
```

## Uninstall

```bash
helm uninstall local-pvc-exporter -n monitoring
```

## Next Steps

- Full configuration reference: [README.md](README.md#%EF%B8%8F-configuration)
- Metrics details: [README.md](README.md#-metrics)
- Troubleshooting permission errors: [README.md](README.md#-troubleshooting)
