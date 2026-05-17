// cmd/power-recovery/main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"power-failure-recovery/internal/checkpoint"
	"power-failure-recovery/internal/config"
	"power-failure-recovery/internal/monitors"
	"power-failure-recovery/internal/scheduler"
	"power-failure-recovery/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// Initialize logging
	if err := logger.Init(cfg.DataDir, cfg.Debug); err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer logger.Close()

	// Print banner
	log.Println("========================================")
	log.Println("Power Failure Recovery Daemon v1.0")
	log.Println("========================================")
	log.Printf("Checkpoint interval: %d seconds", cfg.CheckpointInterval)
	log.Printf("Aggressive interval: %d seconds", cfg.AggressiveInterval)
	log.Printf("Emergency interval: %d seconds", cfg.EmergencyInterval)
	log.Printf("Min battery: %.1f%%", cfg.MinBatteryPercent)
	log.Printf("Fast discharge: %.1f%%/min", cfg.FastDischargeRate)
	log.Printf("Data directory: %s", cfg.DataDir)
	log.Printf("Dry run: %v", cfg.DryRun)
	log.Println("========================================")

	// Handle restore
	if cfg.Restore {
		log.Println("Restoring previous session...")
		mgr := checkpoint.NewManager(cfg.DataDir, cfg.MaxCheckpoints, cfg.DryRun)
		if ckpt, err := mgr.LoadLatest(); err == nil {
			restoreSession(ckpt)
		} else {
			log.Printf("No checkpoint found: %v", err)
		}
		return
	}

	// Initialize components
	batteryMon := monitors.NewBatteryMonitor(cfg.FastDischargeRate, cfg.MinBatteryPercent)
	processMon := monitors.NewProcessMonitor()
	checkpointMgr := checkpoint.NewManager(cfg.DataDir, cfg.MaxCheckpoints, cfg.DryRun)
	sched := scheduler.NewScheduler(cfg.CheckpointInterval, cfg.AggressiveInterval, cfg.EmergencyInterval)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	// Start monitoring loops
	go batteryLoop(batteryMon, checkpointMgr, sched, cfg)
	go checkpointLoop(checkpointMgr, sched, processMon, batteryMon)

	log.Println("Daemon running...")

	// Wait for shutdown
	<-sigChan
	log.Println("Shutdown signal received")

	// Final checkpoint
	saveCheckpoint(checkpointMgr, processMon, batteryMon, "shutdown")

	log.Println("Daemon stopped")
}

func batteryLoop(batteryMon *monitors.BatteryMonitor, checkpointMgr *checkpoint.Manager, sched *scheduler.Scheduler, cfg *config.Config) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sample, err := batteryMon.Read()
		if err != nil {
			log.Printf("Battery read error: %v", err)
			continue
		}

		dischargeInfo := batteryMon.GetDischargeInfo()
		isEmergency := batteryMon.CheckEmergency()
		acOnline := batteryMon.IsACOnline()

		// Log status
		if dischargeInfo.RatePerMinute > 0 {
			log.Printf("Battery: %.1f%% | Discharging: %.1f%%/min | AC: %v | Mode: %s",
				sample.Percent, dischargeInfo.RatePerMinute, acOnline, sched.CurrentMode())
		} else {
			log.Printf("Battery: %.1f%% | AC: %v | Mode: %s",
				sample.Percent, acOnline, sched.CurrentMode())
		}

		// AC loss detection
		if batteryMon.LastACStatus && !acOnline {
			log.Printf("⚠️ AC power lost! Saving emergency checkpoint...")
			saveCheckpoint(checkpointMgr, nil, batteryMon, "ac_loss")
		}
		batteryMon.LastACStatus = acOnline

		// Update scheduler
		sched.Update(dischargeInfo, isEmergency, acOnline)

		// Emergency save
		if isEmergency && !acOnline {
			log.Printf("🚨 CRITICAL: Battery below %.1f%%!", cfg.MinBatteryPercent)
			saveCheckpoint(checkpointMgr, nil, batteryMon, "low_battery")
		}
	}
}

func checkpointLoop(checkpointMgr *checkpoint.Manager, sched *scheduler.Scheduler, processMon *monitors.ProcessMonitor, batteryMon *monitors.BatteryMonitor) {
	for {
		if sched.ShouldCheckpoint() {
			saveCheckpoint(checkpointMgr, processMon, batteryMon, "periodic")
			sched.RecordCheckpoint()
		}
		time.Sleep(1 * time.Second)
	}
}

func saveCheckpoint(mgr *checkpoint.Manager, processMon *monitors.ProcessMonitor, batteryMon *monitors.BatteryMonitor, reason string) {
	sample, err := batteryMon.Read()
	if err != nil {
		log.Printf("Failed to read battery: %v", err)
		return
	}

	dischargeInfo := batteryMon.GetDischargeInfo()
	acOnline := batteryMon.IsACOnline()

	var processes []monitors.Process
	if processMon != nil {
		processes, _ = processMon.GetProcesses()
	}

	ckpt := &checkpoint.Checkpoint{
		Version:        1,
		Timestamp:      time.Now(),
		SaveReason:     reason,
		BatteryPercent: sample.Percent,
		BatteryVoltage: sample.Voltage,
		DischargeRate:  dischargeInfo.RatePerMinute,
		ACOnline:       acOnline,
		Processes:      processes,
	}

	if err := mgr.Save(ckpt); err != nil {
		log.Printf("Failed to save checkpoint: %v", err)
	}
}

func restoreSession(ckpt *checkpoint.Checkpoint) {
	log.Println("========================================")
	log.Printf("Restoring session from %s", ckpt.Timestamp.Format("2006-01-02 15:04:05"))
	log.Printf("Battery at crash: %.1f%%", ckpt.BatteryPercent)
	log.Printf("Processes to restore: %d", len(ckpt.Processes))
	log.Println("========================================")

	for _, proc := range ckpt.Processes {
		if proc.Name == "code" || proc.Name == "firefox" || proc.Name == "chrome" {
			log.Printf("  Would restore: %s", proc.Name)
		}
	}
}
