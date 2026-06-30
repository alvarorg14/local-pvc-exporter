package collector

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/alvarorg/local-pvc-exporter/internal/config"
	"github.com/alvarorg/local-pvc-exporter/internal/discovery"
	"github.com/alvarorg/local-pvc-exporter/internal/diskusage"
	"github.com/alvarorg/local-pvc-exporter/internal/kube"
)

// volumeSample holds scraped metrics for a single volume.
type volumeSample struct {
	labels        prometheus.Labels
	capacityBytes int64
	usedBytes     int64
	inodesUsed    int64
}

// Collector scrapes PVC metrics on an interval and exposes them via Prometheus.
type Collector struct {
	cfg        *config.Config
	discoverer *discovery.Discoverer
	walker     *diskusage.Walker

	mu            sync.RWMutex
	samples       []volumeSample
	lastScrape    time.Time
	scrapeErrors  float64
	scrapeSeconds float64

	capacityGauge   *prometheus.GaugeVec
	usedGauge       *prometheus.GaugeVec
	availableGauge  *prometheus.GaugeVec
	usedRatioGauge  *prometheus.GaugeVec
	inodesUsedGauge *prometheus.GaugeVec

	scrapeDuration prometheus.Gauge
	scrapeErrorsM  prometheus.Gauge
	lastScrapeTS   prometheus.Gauge
}

// New creates a Collector and registers metrics with the provided registry.
func New(cfg *config.Config, client *kube.Client, reg prometheus.Registerer) *Collector {
	discoverer := discovery.New(client, cfg.NodeName, cfg.HostRoot)
	walker := diskusage.New(cfg.DUTimeout)

	suffix := cfg.MetricSuffix()
	prefix := cfg.MetricPrefix

	c := &Collector{
		cfg:        cfg,
		discoverer: discoverer,
		walker:     walker,
	}

	labelNames := []string{
		"persistentvolumeclaim",
		"namespace",
		"persistentvolume",
		"storageclass",
		"node",
		"volume_type",
	}

	c.capacityGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prefix + "_capacity_" + suffix,
		Help: "Declared capacity of the PVC.",
	}, labelNames)

	c.usedGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prefix + "_used_" + suffix,
		Help: "Measured used capacity of the PVC (du-style).",
	}, labelNames)

	c.availableGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prefix + "_available_" + suffix,
		Help: "Available capacity of the PVC (capacity minus used, clamped at 0).",
	}, labelNames)

	c.usedRatioGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prefix + "_used_ratio",
		Help: "Ratio of used to declared capacity (0..1).",
	}, labelNames)

	c.inodesUsedGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: prefix + "_inodes_used",
		Help: "Number of files and directories in the PVC data path.",
	}, labelNames)

	c.scrapeDuration = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: prefix + "_scrape_duration_seconds",
		Help: "Duration of the last scrape in seconds.",
	})

	c.scrapeErrorsM = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: prefix + "_scrape_errors_total",
		Help: "Total number of scrape errors encountered.",
	})

	c.lastScrapeTS = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: prefix + "_last_scrape_timestamp_seconds",
		Help: "Unix timestamp of the last successful scrape.",
	})

	reg.MustRegister(
		c.capacityGauge,
		c.usedGauge,
		c.availableGauge,
		c.usedRatioGauge,
		c.inodesUsedGauge,
		c.scrapeDuration,
		c.scrapeErrorsM,
		c.lastScrapeTS,
	)

	return c
}

// Run starts the background scrape loop until ctx is cancelled.
func (c *Collector) Run(ctx context.Context) {
	c.scrape(ctx)

	ticker := time.NewTicker(c.cfg.ScrapeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.scrape(ctx)
		}
	}
}

func (c *Collector) scrape(ctx context.Context) {
	start := time.Now()
	slog.Info("starting scrape")

	volumes, err := c.discoverer.Discover(ctx)
	if err != nil {
		slog.Error("discovery failed", "error", err)
		c.recordScrapeFailure(start)
		return
	}

	samples := make([]volumeSample, 0, len(volumes))
	sem := make(chan struct{}, c.cfg.DUConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var scrapeErrors int

	for _, vol := range volumes {
		wg.Add(1)
		go func(vol discovery.Volume) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result, err := c.walker.Measure(ctx, vol.HostPath)
			if err != nil {
				slog.Error("du failed",
					"pvc", vol.PVCName,
					"namespace", vol.Namespace,
					"path", vol.HostPath,
					"error", err,
				)
				mu.Lock()
				scrapeErrors++
				mu.Unlock()
				return
			}

			labels := prometheus.Labels{
				"persistentvolumeclaim": vol.PVCName,
				"namespace":               vol.Namespace,
				"persistentvolume":        vol.PVName,
				"storageclass":            vol.StorageClass,
				"node":                    vol.NodeName,
				"volume_type":             string(vol.VolumeType),
			}

			mu.Lock()
			samples = append(samples, volumeSample{
				labels:        labels,
				capacityBytes: vol.CapacityBytes,
				usedBytes:     result.UsedBytes,
				inodesUsed:    result.InodesUsed,
			})
			mu.Unlock()
		}(vol)
	}

	wg.Wait()

	c.updateMetrics(samples, time.Since(start), float64(scrapeErrors))
	slog.Info("scrape complete",
		"volumes", len(samples),
		"errors", scrapeErrors,
		"duration", time.Since(start),
	)
}

func (c *Collector) updateMetrics(samples []volumeSample, duration time.Duration, errors float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.capacityGauge.Reset()
	c.usedGauge.Reset()
	c.availableGauge.Reset()
	c.usedRatioGauge.Reset()
	c.inodesUsedGauge.Reset()

	for _, s := range samples {
		capacity := c.cfg.ConvertBytes(s.capacityBytes)
		used := c.cfg.ConvertBytes(s.usedBytes)
		available := c.cfg.ConvertBytes(maxInt64(0, s.capacityBytes-s.usedBytes))

		c.capacityGauge.With(s.labels).Set(capacity)
		c.usedGauge.With(s.labels).Set(used)
		c.availableGauge.With(s.labels).Set(available)

		var ratio float64
		if s.capacityBytes > 0 {
			ratio = float64(s.usedBytes) / float64(s.capacityBytes)
		}
		c.usedRatioGauge.With(s.labels).Set(ratio)
		c.inodesUsedGauge.With(s.labels).Set(float64(s.inodesUsed))
	}

	c.samples = samples
	c.lastScrape = time.Now()
	c.scrapeSeconds = duration.Seconds()
	c.scrapeErrors += errors

	c.scrapeDuration.Set(c.scrapeSeconds)
	c.scrapeErrorsM.Set(c.scrapeErrors)
	c.lastScrapeTS.Set(float64(c.lastScrape.Unix()))
}

func (c *Collector) recordScrapeFailure(start time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.scrapeErrors++
	c.scrapeSeconds = time.Since(start).Seconds()
	c.scrapeDuration.Set(c.scrapeSeconds)
	c.scrapeErrorsM.Set(c.scrapeErrors)
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// ScrapeOnce performs a single scrape (useful for testing).
func (c *Collector) ScrapeOnce(ctx context.Context) error {
	volumes, err := c.discoverer.Discover(ctx)
	if err != nil {
		return fmt.Errorf("discover: %w", err)
	}
	_ = volumes
	c.scrape(ctx)
	return nil
}
