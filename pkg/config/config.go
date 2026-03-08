// Package config manages probe's configuration loaded from ~/.probe/config.yaml.
// All fields have sensible defaults; the config file is never required.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all probe configuration options.
type Config struct {
	Proxy struct {
		Port           int           `yaml:"port"`            // default 8080
		DashboardPort  int           `yaml:"dashboard_port"`  // default 4041
		StallThreshold time.Duration `yaml:"stall_threshold"` // default 500ms
	} `yaml:"proxy"`
	Alerts struct {
		CostThreshold    float64       `yaml:"cost_threshold"`    // 0 = disabled
		LatencyThreshold time.Duration `yaml:"latency_threshold"` // 0 = disabled
		AlertOnError     bool          `yaml:"alert_on_error"`
	} `yaml:"alerts"`
	Storage struct {
		Persist        bool `yaml:"persist"`           // default false
		RetentionDays  int  `yaml:"retention_days"`    // default 7
		RingBufferSize int  `yaml:"ring_buffer_size"`  // default 1000
	} `yaml:"storage"`
	Pricing struct {
		Custom map[string]CustomPricing `yaml:"custom"` // model -> pricing overrides
	} `yaml:"pricing"`
}

// CustomPricing holds user-defined per-1M-token pricing for a model.
type CustomPricing struct {
	InputPer1M  float64 `yaml:"input"`
	OutputPer1M float64 `yaml:"output"`
}

// Default returns a Config with all default values filled in.
func Default() *Config {
	cfg := &Config{}
	cfg.Proxy.Port = 8080
	cfg.Proxy.DashboardPort = 4041
	cfg.Proxy.StallThreshold = 500 * time.Millisecond
	cfg.Storage.RetentionDays = 7
	cfg.Storage.RingBufferSize = 1000
	cfg.Pricing.Custom = make(map[string]CustomPricing)
	return cfg
}

// Load reads ~/.probe/config.yaml, returning defaults if absent.
// If the file exists but cannot be parsed, an error is returned.
func Load() (*Config, error) {
	cfg := Default()

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg, fmt.Errorf("config: cannot determine home directory: %w", err)
	}

	path := filepath.Join(home, ".probe", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file — use defaults silently.
			return cfg, nil
		}
		return cfg, fmt.Errorf("config: reading %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, fmt.Errorf("config: parsing %s: %w", path, err)
	}

	// Re-apply defaults for zero values so callers never see 0 for required fields.
	if cfg.Proxy.Port == 0 {
		cfg.Proxy.Port = 8080
	}
	if cfg.Proxy.DashboardPort == 0 {
		cfg.Proxy.DashboardPort = 4041
	}
	if cfg.Proxy.StallThreshold == 0 {
		cfg.Proxy.StallThreshold = 500 * time.Millisecond
	}
	if cfg.Storage.RetentionDays == 0 {
		cfg.Storage.RetentionDays = 7
	}
	if cfg.Storage.RingBufferSize == 0 {
		cfg.Storage.RingBufferSize = 1000
	}
	if cfg.Pricing.Custom == nil {
		cfg.Pricing.Custom = make(map[string]CustomPricing)
	}

	return cfg, nil
}
