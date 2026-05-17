// internal/monitors/battery.go
package monitors

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// BatterySample represents a single battery reading
type BatterySample struct {
	Timestamp time.Time `json:"timestamp"`
	Percent   float64   `json:"percent"`
	Voltage   int       `json:"voltage_mv"`
	Current   int       `json:"current_ma"`
	Status    string    `json:"status"`
}

// DischargeInfo contains discharge rate calculations
type DischargeInfo struct {
	RatePerMinute float64 `json:"rate_per_minute"`
	IsFast        bool    `json:"is_fast"`
	TimeRemaining int64   `json:"time_remaining_secs"`
}

// BatteryMonitor monitors battery status
type BatteryMonitor struct {
	samples                []BatterySample
	fastDischargeThreshold float64
	minBatteryThreshold    float64
	LastACStatus           bool
	batteryPath            string
}

// NewBatteryMonitor creates a new battery monitor
func NewBatteryMonitor(fastDischargeThreshold, minBatteryThreshold float64) *BatteryMonitor {
	// Find battery path
	batteryPath := ""
	for _, path := range []string{"/sys/class/power_supply/BAT0", "/sys/class/power_supply/BAT1"} {
		if _, err := os.Stat(path); err == nil {
			batteryPath = path
			break
		}
	}

	return &BatteryMonitor{
		samples:                make([]BatterySample, 0),
		fastDischargeThreshold: fastDischargeThreshold,
		minBatteryThreshold:    minBatteryThreshold,
		LastACStatus:           true,
		batteryPath:            batteryPath,
	}
}

// Read reads current battery status from sysfs
func (bm *BatteryMonitor) Read() (*BatterySample, error) {
	if bm.batteryPath == "" {
		return nil, fmt.Errorf("no battery found")
	}

	sample := &BatterySample{
		Timestamp: time.Now(),
	}

	// Read capacity (percentage)
	if data, err := os.ReadFile(bm.batteryPath + "/capacity"); err == nil {
		sample.Percent, _ = strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	}

	// Read status
	if data, err := os.ReadFile(bm.batteryPath + "/status"); err == nil {
		sample.Status = strings.TrimSpace(string(data))
	}

	// Read voltage (microvolts -> millivolts)
	if data, err := os.ReadFile(bm.batteryPath + "/voltage_now"); err == nil {
		if uv, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			sample.Voltage = int(uv / 1000)
		}
	}

	// Read current (microamps -> milliamps)
	if data, err := os.ReadFile(bm.batteryPath + "/current_now"); err == nil {
		if ua, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			sample.Current = int(ua / 1000)
		}
	}

	// Store sample
	bm.samples = append(bm.samples, *sample)

	// Keep last 30 samples (5 minutes at 10s intervals)
	if len(bm.samples) > 30 {
		bm.samples = bm.samples[1:]
	}

	return sample, nil
}

// IsACOnline determines AC status from battery state - FIXED!
// If battery is Discharging, AC must be false (unplugged)
// If battery is Charging or Full, AC must be true (plugged in)
func (bm *BatteryMonitor) IsACOnline() bool {
	if len(bm.samples) == 0 {
		return true // Assume AC is connected initially
	}

	latest := bm.samples[len(bm.samples)-1]

	// This is the reliable method:
	// Battery status tells us directly if AC is connected
	switch latest.Status {
	case "Discharging":
		return false // AC is NOT connected (running on battery)
	case "Charging", "Full":
		return true // AC IS connected
	default:
		// Fallback: if status unknown, assume AC connected
		return true
	}
}

// GetDischargeInfo calculates discharge rate
func (bm *BatteryMonitor) GetDischargeInfo() *DischargeInfo {
	if len(bm.samples) < 2 {
		return &DischargeInfo{RatePerMinute: 0, IsFast: false}
	}

	// Filter only discharging samples
	discharging := make([]BatterySample, 0)
	for _, s := range bm.samples {
		if s.Status == "Discharging" {
			discharging = append(discharging, s)
		}
	}

	if len(discharging) < 2 {
		return &DischargeInfo{RatePerMinute: 0, IsFast: false}
	}

	first := discharging[0]
	last := discharging[len(discharging)-1]

	timeDiff := last.Timestamp.Sub(first.Timestamp).Minutes()
	percentDiff := first.Percent - last.Percent

	if timeDiff <= 0 || percentDiff <= 0 {
		return &DischargeInfo{RatePerMinute: 0, IsFast: false}
	}

	ratePerMinute := percentDiff / timeDiff
	isFast := ratePerMinute >= bm.fastDischargeThreshold

	var timeRemaining int64 = 0
	if ratePerMinute > 0 {
		timeRemaining = int64(last.Percent / ratePerMinute * 60)
	}

	return &DischargeInfo{
		RatePerMinute: ratePerMinute,
		IsFast:        isFast,
		TimeRemaining: timeRemaining,
	}
}

// CheckEmergency checks if emergency conditions are met
func (bm *BatteryMonitor) CheckEmergency() bool {
	if len(bm.samples) == 0 {
		return false
	}

	latest := bm.samples[len(bm.samples)-1]
	acOnline := bm.IsACOnline()

	return !acOnline && latest.Percent <= bm.minBatteryThreshold
}
