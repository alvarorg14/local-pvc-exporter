# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /out/local-pvc-exporter \
    ./cmd/local-pvc-exporter

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /out/local-pvc-exporter /local-pvc-exporter

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/local-pvc-exporter"]
