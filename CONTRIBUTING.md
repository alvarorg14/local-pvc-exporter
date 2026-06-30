# Contributing to local-pvc-exporter

Thank you for your interest in contributing!

## Getting started

1. Fork the repository and clone it locally.
2. Install Go 1.23+.
3. Run `go mod download`.
4. Make your changes and add tests where appropriate.
5. Run `make test` and `make lint` before opening a pull request.

## Development workflow

```bash
# Run tests
make test

# Build binary
make build

# Lint (requires golangci-lint)
make lint

# Build container image
make docker
```

## Pull requests

- Keep changes focused and well-scoped.
- Update documentation when behavior or configuration changes.
- Use conventional commit messages when possible (e.g. `feat:`, `fix:`, `docs:`).
- Ensure CI passes before requesting review.

## Reporting issues

Please include:

- Kubernetes distribution and version (e.g. k3s v1.31)
- PV volume type (`hostPath` or `local`)
- Exporter configuration (metric prefix, scrape interval, unit)
- Relevant logs from the exporter pod
- Example PV/PVC YAML (redact secrets)

## Code of conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.
