// internal/config/config.go
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	// Timing
	CheckpointInterval int `json:"checkpoint_interval_secs"`
	AggressiveInterval int `json:"aggressive_interval_secs"`
	EmergencyInterval  int `json:"emergency_interval_secs"`

	// Battery thresholds
	MinBatteryPercent float64 `json:"min_battery_percent"`
	FastDischargeRate float64 `json:"fast_discharge_rate_per_min"`

	// Storage
	DataDir        string `json:"data_dir"`
	MaxCheckpoints int    `json:"max_checkpoints"`

	// Runtime
	DryRun  bool
	Debug   bool
	Restore bool
	SaveNow bool
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		CheckpointInterval: 60,
		AggressiveInterval: 10,
		EmergencyInterval:  1,
		MinBatteryPercent:  5.0,
		FastDischargeRate:  15.0,
		DataDir:            "/var/lib/power-failure-recovery",
		MaxCheckpoints:     10,
		DryRun:             false,
		Debug:              false,
		Restore:            false,
		SaveNow:            false,
	}
}

// Load loads configuration from CLI flags and config file
func Load() (*Config, error) {
	cfg := DefaultConfig()

	// Parse CLI flags
	interval := flag.Int("interval", cfg.CheckpointInterval, "Checkpoint interval (seconds)")
	aggressive := flag.Int("aggressive-interval", cfg.AggressiveInterval, "Aggressive interval (seconds)")
	emergency := flag.Int("emergency-interval", cfg.EmergencyInterval, "Emergency interval (seconds)")
	minBattery := flag.Float64("min-battery", cfg.MinBatteryPercent, "Minimum battery percentage")
	fastDischarge := flag.Float64("fast-discharge", cfg.FastDischargeRate, "Fast discharge rate (%/min)")
	dataDir := flag.String("data-dir", cfg.DataDir, "Data directory")
	maxCheckpoints := flag.Int("max-checkpoints", cfg.MaxCheckpoints, "Maximum checkpoints to keep")
	dryRun := flag.Bool("dry-run", false, "Don't actually save")
	debug := flag.Bool("debug", false, "Enable debug logging")
	restore := flag.Bool("restore", false, "Restore previous session")
	saveNow := flag.Bool("save-now", false, "Save a checkpoint immediately and exit")

	flag.Parse()

	// Try to load config file
	configPaths := []string{
		"/etc/power-failure-recovery/config.json",
		filepath.Join(*dataDir, "config.json"),
		"./configs/default.json",
	}

	for _, path := range configPaths {
		if data, err := os.ReadFile(path); err == nil {
			var fileConfig Config
			if err := json.Unmarshal(data, &fileConfig); err == nil {
				// Merge file config (CLI overrides)
				if fileConfig.CheckpointInterval != 0 {
					cfg.CheckpointInterval = fileConfig.CheckpointInterval
				}
				if fileConfig.AggressiveInterval != 0 {
					cfg.AggressiveInterval = fileConfig.AggressiveInterval
				}
				if fileConfig.EmergencyInterval != 0 {
					cfg.EmergencyInterval = fileConfig.EmergencyInterval
				}
				if fileConfig.MinBatteryPercent != 0 {
					cfg.MinBatteryPercent = fileConfig.MinBatteryPercent
				}
				if fileConfig.FastDischargeRate != 0 {
					cfg.FastDischargeRate = fileConfig.FastDischargeRate
				}
				if fileConfig.DataDir != "" {
					cfg.DataDir = fileConfig.DataDir
				}
				if fileConfig.MaxCheckpoints != 0 {
					cfg.MaxCheckpoints = fileConfig.MaxCheckpoints
				}
			}
			break
		}
	}

	// CLI overrides
	cfg.CheckpointInterval = *interval
	cfg.AggressiveInterval = *aggressive
	cfg.EmergencyInterval = *emergency
	cfg.MinBatteryPercent = *minBattery
	cfg.FastDischargeRate = *fastDischarge
	cfg.DataDir = *dataDir
	cfg.MaxCheckpoints = *maxCheckpoints
	cfg.DryRun = *dryRun
	cfg.Debug = *debug
	cfg.Restore = *restore
	cfg.SaveNow = *saveNow

	// Create data directory
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.CheckpointInterval < 1 {
		return fmt.Errorf("checkpoint interval must be at least 1 second")
	}
	if c.MinBatteryPercent < 0 || c.MinBatteryPercent > 100 {
		return fmt.Errorf("min battery percent must be between 0 and 100")
	}
	if c.FastDischargeRate <= 0 {
		return fmt.Errorf("fast discharge rate must be positive")
	}
	return nil
}
