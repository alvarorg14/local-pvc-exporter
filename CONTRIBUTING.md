# Contributing to local-pvc-exporter

Thank you for your interest in contributing!

## Getting started

1. Fork the repository and clone it locally.
2. Install Go 1.26+.
3. Run `go mod download`.
4. Make your changes and add tests where appropriate.
5. Run `make test` and `make lint` before opening a pull request.

Run `make help` to see all available development targets.

## Development workflow

```bash
# List all targets
make help

# Run tests
make test

# Run tests with coverage
make cover

# Build binary
make build

# Lint (requires golangci-lint)
make lint

# Format code
make fmt

# Run locally (requires kubeconfig)
make run

# Build container image
make docker

# Lint Helm chart
make helm-lint
```

For architecture details and AI assistant guidelines, see [AGENTS.md](AGENTS.md).

## Pull requests

- Keep changes focused and well-scoped.
- Update documentation when behavior or configuration changes.
- Use conventional commit messages when possible (e.g. `feat:`, `fix:`, `docs:`).
- Ensure CI passes before requesting review.
- Add exactly one policy label to your PR: `breaking-change`, `feature`, `enhancement`, `bug`, `dependencies`, `documentation`, `deprecations`, or `ci`.

## Reporting issues

Please include:

- Kubernetes distribution and version (e.g. k3s v1.31)
- PV volume type (`hostPath` or `local`)
- Exporter configuration (metric prefix, scrape interval, unit)
- Relevant logs from the exporter pod
- Example PV/PVC YAML (redact secrets)

## Security

If you discover a security vulnerability, please report it via a [private GitHub security advisory](https://github.com/alvarorg14/local-pvc-exporter/security/advisories/new). See [SECURITY.md](SECURITY.md) for details.

## Code of conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.
