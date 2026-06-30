package config

import (
	"testing"
	"time"
)

func TestParseUnit(t *testing.T) {
	tests := []struct {
		input   string
		want    Unit
		wantErr bool
	}{
		{"bytes", UnitBytes, false},
		{"BYTES", UnitBytes, false},
		{"kib", UnitKiB, false},
		{"mib", UnitMiB, false},
		{"gib", UnitGiB, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		got, err := ParseUnit(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("ParseUnit(%q) expected error", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseUnit(%q) unexpected error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("ParseUnit(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestConvertBytes(t *testing.T) {
	cfg := &Config{Unit: UnitKiB}
	if got := cfg.ConvertBytes(2048); got != 2 {
		t.Errorf("ConvertBytes(2048) with kib = %f, want 2", got)
	}

	cfg.Unit = UnitMiB
	if got := cfg.ConvertBytes(2 * 1024 * 1024); got != 2 {
		t.Errorf("ConvertBytes with mib = %f, want 2", got)
	}

	cfg.Unit = UnitBytes
	if got := cfg.ConvertBytes(100); got != 100 {
		t.Errorf("ConvertBytes with bytes = %f, want 100", got)
	}
}

func TestMetricSuffix(t *testing.T) {
	tests := []struct {
		unit Unit
		want string
	}{
		{UnitBytes, "bytes"},
		{UnitKiB, "kib"},
		{UnitMiB, "mib"},
		{UnitGiB, "gib"},
	}

	for _, tt := range tests {
		cfg := &Config{Unit: tt.unit}
		if got := cfg.MetricSuffix(); got != tt.want {
			t.Errorf("MetricSuffix() for %q = %q, want %q", tt.unit, got, tt.want)
		}
	}
}

func TestEnvDuration(t *testing.T) {
	t.Setenv("TEST_DURATION", "30s")
	if got := envDuration("TEST_DURATION", time.Minute); got != 30*time.Second {
		t.Errorf("envDuration = %v, want 30s", got)
	}
}
