# AGENTS.md - AI Assistant Guidelines

This document provides context and guidelines for AI coding assistants working on the local-pvc-exporter project.

## Project Overview

**local-pvc-exporter** is a Prometheus exporter that exposes **per-PVC storage metrics** on Kubernetes clusters where `kubelet_volume_stats_*` metrics are missing or inaccurate.

### Key Purpose

- Measure used capacity for `hostPath` and `local` PersistentVolumes via du-style directory walks
- Expose Prometheus metrics with standard Kubernetes labels (`persistentvolumeclaim`, `namespace`, `persistentvolume`, `storageclass`, `node`, `volume_type`)
- Run as a DaemonSet on each node, discovering volumes bound to that node
- Fill the gap for k3s, edge, and other deployments where kubelet volume stats are absent or filesystem-level only

## Architecture

### Backend (Go)

- **Location**: `cmd/local-pvc-exporter/` (main entry), `internal/` (core logic)
- **Module**: `github.com/alvarorg14/local-pvc-exporter`
- **Go version**: 1.23+
- **Key dependencies**: `prometheus/client_golang`, `k8s.io/client-go`

### Key Components

| Package | Path | Responsibility |
|---------|------|----------------|
| `config` | `internal/config/` | Flags/env parsing, unit conversion (`bytes`, `kib`, `mib`, `gib`) |
| `collector` | `internal/collector/` | Scrape loop, Prometheus metric registration, concurrent du orchestration |
| `discovery` | `internal/discovery/` | Discover hostPath/local PVs bound to PVCs on this node |
| `diskusage` | `internal/diskusage/` | Du-style directory walk (inode dedup, single-filesystem boundary), statfs inodes |
| `kube` | `internal/kube/` | Thin Kubernetes API client wrapper |

### Helm Chart

- **Location**: `charts/local-pvc-exporter/`
- Deploys DaemonSet, Service, ServiceAccount, RBAC, optional ServiceMonitor
- Default security: root + `CAP_DAC_READ_SEARCH`, read-only host mount, distroless image

## Data Flow

```
1. DaemonSet pod starts on node with NODE_NAME set
2. Discoverer lists PVs/PVCs/StorageClasses from K8s API
3. Filters to hostPath/local volumes on this node
4. Walker measures each volume path under HOST_ROOT (du-style)
5. Collector caches metrics; Prometheus scrapes /metrics
```

## Configuration

| Flag / Env | Default | Description |
|------------|---------|-------------|
| `--metric-prefix` / `METRIC_PREFIX` | `local_pvc` | Prefix for all metrics |
| `--scrape-interval` / `SCRAPE_INTERVAL` | `5m` | Interval between PVC scans |
| `--unit` / `UNIT` | `bytes` | Output unit: `bytes`, `kib`, `mib`, `gib` |
| `--listen-address` / `LISTEN_ADDRESS` | `:8080` | HTTP listen address |
| `--host-root` / `HOST_ROOT` | `/host` | Host filesystem mount inside pod |
| `--node-name` / `NODE_NAME` | **required** | Node name |
| `--du-concurrency` / `DU_CONCURRENCY` | `4` | Max concurrent du operations |
| `--du-timeout` / `DU_TIMEOUT` | `10m` | Per-volume du timeout |
| `--kubeconfig` / `KUBECONFIG` | *(empty)* | Kubeconfig path (in-cluster if empty) |

## Key Files

| File | Purpose |
|------|---------|
| `cmd/local-pvc-exporter/main.go` | Entry point: config load, HTTP server, graceful shutdown |
| `internal/collector/collector.go` | Metric definitions, scrape loop, gauge updates |
| `internal/discovery/discovery.go` | PV/PVC discovery and node matching |
| `internal/discovery/path.go` | Path resolution under host root |
| `internal/diskusage/diskusage.go` | Directory walk and inode measurement |
| `internal/config/config.go` | Flag/env parsing and validation |
| `internal/kube/client.go` | Kubernetes client wrapper |
| `charts/local-pvc-exporter/values.yaml` | Helm defaults |
| `.golangci.yml` | Linter configuration |
| `.goreleaser.yaml` | Release and container image publishing |

## Development Workflow

Use `make` targets as the canonical development commands:

```bash
make help      # List all targets
make test      # go test -race ./...
make cover     # go test -race -coverprofile=coverage.out ./...
make build     # go build -o bin/local-pvc-exporter ./cmd/local-pvc-exporter
make run       # Run locally with NODE_NAME from kubectl
make lint      # golangci-lint run
make fmt       # gofmt + go fmt
make tidy      # go mod tidy
make docker    # docker build (VERSION=dev)
make helm-lint # helm lint + helm template
```

### Prerequisites

- Go 1.23+
- [golangci-lint](https://golangci-lint.run/) v2.x (CI uses v2.12.2)
- Helm 3 (for chart validation)
- kubectl (for `make run`)

## Testing Conventions

- Standard library `testing` only (no testify)
- Table-driven tests where appropriate
- Test files live beside source: `*_test.go`
- Existing tests:
  - `internal/config/config_test.go` — unit parsing and conversion
  - `internal/discovery/discovery_test.go` — node matching helpers
  - `internal/diskusage/diskusage_test.go` — directory walk with temp dirs
- Run: `make test` or `go test -race ./...`

## CI & Release

### CI (`.github/workflows/ci.yml`)

Triggers on push/PR to `main`:

| Job | Command |
|-----|---------|
| test | `go mod verify`, `go test -race -coverprofile=coverage.out ./...`, `go build` |
| lint | golangci-lint v2.12.2 |
| vuln | `govulncheck ./...` |
| helm | `helm lint`, `helm template`, `helm package` |
| goreleaser | `goreleaser check`, `goreleaser release --snapshot --clean` |

### Dependency updates (Renovate)

[Renovate](https://docs.renovatebot.com/) is configured in [`renovate.json`](renovate.json) to propose updates for Go modules, Docker base images, and GitHub Actions. Pull requests are labeled `dependencies` or `github-actions` to satisfy the PR policy below. Install the [Renovate GitHub App](https://github.com/apps/renovate) on the repository to enable automated update PRs.

### PR Policy (`.github/workflows/pr-policy.yml`)

Pull requests must carry **exactly one** label:

- `breaking-change`, `feature`, `enhancement`, `bug`, `dependencies`, `documentation`, `deprecations`, `github-actions`

### Release (`.github/workflows/release.yml`)

On GitHub release publish:

- GoReleaser builds binaries and pushes `ghcr.io/alvarorg14/local-pvc-exporter` (amd64 + arm64)
- Helm chart packaged and pushed to `oci://ghcr.io/alvarorg14/charts`

## Quality Assurance Requirements

**Every change MUST pass these checks before completion:**

### 1. Linting (MANDATORY)

- Run `make lint` — must pass with 0 errors
- Linters enabled: `errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`
- Zero tolerance: `max-issues-per-linter: 0`

### 2. Build (MANDATORY)

- `make build` must succeed
- Never commit code that doesn't build

### 3. Testing (MANDATORY)

- `make test` must pass
- Do NOT skip or delete tests to make them pass
- Add tests for new features when appropriate

### 4. Documentation (MANDATORY)

- Update `README.md` for user-facing changes
- Update `QUICKSTART.md` if install/verify steps change
- Update `AGENTS.md` for architectural changes
- Update Helm chart README if values change

### Pre-Commit Checklist

- [ ] `make lint` passes (0 errors)
- [ ] `make build` succeeds
- [ ] `make test` passes
- [ ] Documentation updated where needed
- [ ] PR has exactly one policy label
- [ ] Conventional commit message used when possible

## Code Style

- **Go**: Follow standard Go conventions, use `gofmt`
- **Error handling**: Always handle errors; use structured `slog` for logging
- **Context**: Use `context.Context` with timeouts for long-running operations (du walks)
- **Concurrency**: Respect `du-concurrency` and `du-timeout` settings in collector

## Important Warnings

**DO NOT**:

- Commit code that doesn't build or pass tests
- Skip linting, testing, or documentation updates
- Open public issues for security vulnerabilities (use private advisories)
- Grant write access or unnecessary capabilities in Helm values
- Mark tasks complete without verification

**DO**:

- Use `make` targets for development
- Add tests for new logic in `internal/`
- Keep RBAC read-only (get/list/watch only)
- Update docs when behavior or configuration changes
- Use conventional commits (`feat:`, `fix:`, `docs:`, etc.)

## Related Documentation

- [README.md](README.md) — User-facing overview and reference
- [QUICKSTART.md](QUICKSTART.md) — Get running in minutes
- [CONTRIBUTING.md](CONTRIBUTING.md) — Contributor workflow
- [SECURITY.md](SECURITY.md) — Security policy and vulnerability reporting
