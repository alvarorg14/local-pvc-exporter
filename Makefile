VERSION ?= dev

.PHONY: help test cover build run lint fmt tidy docker helm-lint

help: ## Show available targets
	@grep -E '^[a-zA-Z0-9_-]+:.*##' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "}; {printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2}'

test: ## Run tests with race detector
	go test -race ./...

cover: ## Run tests with coverage profile
	go test -race -coverprofile=coverage.out ./...

build: ## Build binary to bin/local-pvc-exporter
	go build -o bin/local-pvc-exporter ./cmd/local-pvc-exporter

run: ## Run locally against the current kubeconfig
	NODE_NAME=$$(kubectl get nodes -o jsonpath='{.items[0].metadata.name}') \
		go run ./cmd/local-pvc-exporter --host-root=/

lint: ## Run golangci-lint
	golangci-lint run

fmt: ## Format Go source files
	gofmt -w .
	go fmt ./...

tidy: ## Tidy Go module dependencies
	go mod tidy

docker: ## Build container image (VERSION=dev by default)
	docker build --build-arg VERSION=$(VERSION) -t local-pvc-exporter:$(VERSION) .

helm-lint: ## Lint and template the Helm chart
	helm lint charts/local-pvc-exporter
	helm template test charts/local-pvc-exporter > /dev/null
