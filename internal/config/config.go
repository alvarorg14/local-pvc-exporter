package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Unit represents the output unit for capacity metrics.
type Unit string

const (
	UnitBytes Unit = "bytes"
	UnitKiB   Unit = "kib"
	UnitMiB   Unit = "mib"
	UnitGiB   Unit = "gib"
)

// Config holds all exporter configuration.
type Config struct {
	MetricPrefix   string
	ScrapeInterval time.Duration
	Unit           Unit
	ListenAddress  string
	HostRoot       string
	Kubeconfig     string
	NodeName       string
	DUConcurrency  int
	DUTimeout      time.Duration
}

// MetricSuffix returns the suffix used in metric names for the configured unit.
func (c *Config) MetricSuffix() string {
	switch c.Unit {
	case UnitKiB:
		return "kib"
	case UnitMiB:
		return "mib"
	case UnitGiB:
		return "gib"
	default:
		return "bytes"
	}
}

// ConvertBytes converts a byte value to the configured output unit.
func (c *Config) ConvertBytes(bytes int64) float64 {
	switch c.Unit {
	case UnitKiB:
		return float64(bytes) / 1024
	case UnitMiB:
		return float64(bytes) / (1024 * 1024)
	case UnitGiB:
		return float64(bytes) / (1024 * 1024 * 1024)
	default:
		return float64(bytes)
	}
}

// ParseUnit parses a unit string.
func ParseUnit(s string) (Unit, error) {
	switch Unit(strings.ToLower(strings.TrimSpace(s))) {
	case UnitBytes, UnitKiB, UnitMiB, UnitGiB:
		return Unit(strings.ToLower(strings.TrimSpace(s))), nil
	default:
		return "", fmt.Errorf("invalid unit %q: must be one of bytes, kib, mib, gib", s)
	}
}

// Load parses configuration from flags and environment variables.
// Environment variables take precedence over flag defaults when set.
func Load() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.MetricPrefix, "metric-prefix", envString("METRIC_PREFIX", "local_pvc"), "Prefix for all exported metrics")
	flag.DurationVar(&cfg.ScrapeInterval, "scrape-interval", envDuration("SCRAPE_INTERVAL", 5*time.Minute), "Interval between PVC capacity scrapes")
	unitDefault := envString("UNIT", "bytes")
	var unitStr string
	flag.StringVar(&unitStr, "unit", unitDefault, "Output unit for capacity metrics (bytes, kib, mib, gib)")
	flag.StringVar(&cfg.ListenAddress, "listen-address", envString("LISTEN_ADDRESS", ":8080"), "Address to listen on for HTTP requests")
	flag.StringVar(&cfg.HostRoot, "host-root", envString("HOST_ROOT", "/host"), "Mount point of the host filesystem inside the pod")
	flag.StringVar(&cfg.Kubeconfig, "kubeconfig", envString("KUBECONFIG", ""), "Path to kubeconfig file (empty for in-cluster config)")
	flag.StringVar(&cfg.NodeName, "node-name", envString("NODE_NAME", ""), "Name of the node this exporter runs on")
	flag.IntVar(&cfg.DUConcurrency, "du-concurrency", envInt("DU_CONCURRENCY", 4), "Maximum concurrent du operations")
	flag.DurationVar(&cfg.DUTimeout, "du-timeout", envDuration("DU_TIMEOUT", 10*time.Minute), "Timeout for a single du operation")

	flag.Parse()

	unit, err := ParseUnit(unitStr)
	if err != nil {
		return nil, err
	}
	cfg.Unit = unit

	if cfg.NodeName == "" {
		return nil, errors.New("node-name is required (set --node-name or NODE_NAME)")
	}
	if cfg.MetricPrefix == "" {
		return nil, errors.New("metric-prefix cannot be empty")
	}
	if cfg.DUConcurrency < 1 {
		return nil, errors.New("du-concurrency must be at least 1")
	}
	if cfg.ScrapeInterval < time.Second {
		return nil, errors.New("scrape-interval must be at least 1s")
	}

	return cfg, nil
}

func envString(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
