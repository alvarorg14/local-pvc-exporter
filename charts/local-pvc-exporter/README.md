# local-pvc-exporter

Prometheus exporter for hostPath and local PV capacity metrics per PVC

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

## Installing

```bash
helm install local-pvc-exporter oci://ghcr.io/alvarorg14/charts/local-pvc-exporter \
  --namespace monitoring \
  --create-namespace
```

Enable Prometheus Operator scraping:

```bash
helm install local-pvc-exporter oci://ghcr.io/alvarorg14/charts/local-pvc-exporter \
  --namespace monitoring \
  --create-namespace \
  --set serviceMonitor.enabled=true
```

For local development, install from the chart source:

```bash
helm install local-pvc-exporter ./charts/local-pvc-exporter \
  --namespace monitoring \
  --create-namespace
```

## Values
| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity rules for pod assignment |
| duConcurrency | int | `4` | Maximum concurrent du operations per scrape (maps to `--du-concurrency`) |
| duTimeout | string | `"10m"` | Timeout for a single du operation (maps to `--du-timeout`) |
| extraArgs | list | `[]` | Extra command-line arguments passed to the exporter |
| extraEnv | list | `[]` | Extra environment variables for the exporter container |
| fullnameOverride | string | `""` | Override the full release name used for all resources |
| hostRoot | string | `"/host"` | Mount point of the host filesystem inside the pod (maps to `--host-root`) |
| hostRootMountPath | string | `"/"` | Host path mounted read-only at hostRoot for PVC directory traversal |
| image.pullPolicy | string | `"IfNotPresent"` | Image pull policy |
| image.repository | string | `"ghcr.io/alvarorg14/local-pvc-exporter"` | Container image repository |
| image.tag | string | `""` | Image tag. Defaults to the chart appVersion when empty |
| imagePullSecrets | list | `[]` | Secrets for pulling images from private registries |
| listenAddress | string | `":8080"` | HTTP listen address for metrics and health endpoints (maps to `--listen-address`) |
| metricPrefix | string | `"local_pvc"` | Prefix for all exported metrics (maps to `--metric-prefix`) |
| nameOverride | string | `""` | Override the chart name used in labels and resource names |
| nodeSelector | object | `{}` | Node labels for pod assignment |
| podAnnotations | object | `{}` | Annotations to add to exporter pods |
| podLabels | object | `{}` | Labels to add to exporter pods |
| podSecurityContext | object | `{"runAsGroup":0,"runAsUser":0}` | Pod-level security context. Default runs as root to traverse PVC data directories with restrictive permissions |
| rbac.create | bool | `true` | Create ClusterRole and ClusterRoleBinding for PV/PVC/node discovery |
| replicaCount | int | `1` | Number of DaemonSet pods. Ignored: the chart deploys a DaemonSet, one pod per node. |
| resources | object | `{"limits":{"memory":"128Mi"},"requests":{"cpu":"50m","memory":"128Mi"}}` | CPU and memory resource requests and limits for the exporter container |
| scrapeInterval | string | `"5m"` | Interval between PVC capacity scrapes (maps to `--scrape-interval`) |
| securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"add":["DAC_READ_SEARCH"],"drop":["ALL"]},"readOnlyRootFilesystem":true}` | Container-level security context. Adds CAP_DAC_READ_SEARCH for read-only traversal of host paths |
| service.annotations | object | `{}` | Annotations to add to the Service |
| service.port | int | `8080` | Service port exposed for Prometheus scraping |
| service.type | string | `"ClusterIP"` | Kubernetes Service type |
| serviceAccount.annotations | object | `{}` | Annotations to add to the ServiceAccount |
| serviceAccount.create | bool | `true` | Create a dedicated ServiceAccount for the exporter |
| serviceAccount.name | string | `""` | ServiceAccount name. Generated from the release fullname when empty and create is true |
| serviceMonitor.enabled | bool | `false` | Create a Prometheus Operator ServiceMonitor resource |
| serviceMonitor.interval | string | `"30s"` | Scrape interval for the ServiceMonitor |
| serviceMonitor.labels | object | `{}` | Additional labels for the ServiceMonitor (e.g. release label for Prometheus Operator) |
| serviceMonitor.namespace | string | `""` | Namespace for the ServiceMonitor. Defaults to the release namespace when empty |
| serviceMonitor.scrapeTimeout | string | `"10s"` | Scrape timeout for the ServiceMonitor |
| tolerations | list | `[{"effect":"NoSchedule","key":"node-role.kubernetes.io/control-plane","operator":"Exists"},{"effect":"NoSchedule","key":"node-role.kubernetes.io/master","operator":"Exists"}]` | Tolerations for pod assignment. Default tolerates control-plane/master taints so the DaemonSet runs on all nodes |
| unit | string | `"bytes"` | Output unit for capacity metrics: `bytes`, `kib`, `mib`, `gib` (maps to `--unit`) |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
