// internal/checkpoint/types.go
package checkpoint

import (
	"time"

	"power-failure-recovery/internal/monitors"
)

// Checkpoint represents a saved system state
type Checkpoint struct {
	Version        int                     `json:"version"`
	Timestamp      time.Time               `json:"timestamp"`
	SaveReason     string                  `json:"save_reason"`
	BatteryPercent float64                 `json:"battery_percent"`
	BatteryVoltage int                     `json:"battery_voltage_mv"`
	DischargeRate  float64                 `json:"discharge_rate_per_min"`
	ACOnline       bool                    `json:"ac_online"`
	Processes      []monitors.Process      `json:"processes"`
	ActiveWindows  []monitors.ActiveWindow `json:"active_windows,omitempty"`
}
