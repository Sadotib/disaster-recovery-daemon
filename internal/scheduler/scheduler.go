// internal/scheduler/scheduler.go
package scheduler

import (
	"log"
	"time"

	"power-failure-recovery/internal/monitors"
)

// Mode represents the current saving mode
type Mode string

const (
	ModeNormal     Mode = "normal"
	ModeAggressive Mode = "aggressive"
	ModeEmergency  Mode = "emergency"
)

// Scheduler adapts checkpoint frequency based on conditions
type Scheduler struct {
	normalInterval     time.Duration
	aggressiveInterval time.Duration
	emergencyInterval  time.Duration

	currentMode              Mode
	lastCheckpoint           time.Time
	consecutiveFastDischarge int
}

// NewScheduler creates a new adaptive scheduler
func NewScheduler(normalSec, aggressiveSec, emergencySec int) *Scheduler {
	return &Scheduler{
		normalInterval:     time.Duration(normalSec) * time.Second,
		aggressiveInterval: time.Duration(aggressiveSec) * time.Second,
		emergencyInterval:  time.Duration(emergencySec) * time.Second,
		currentMode:        ModeNormal,
		lastCheckpoint:     time.Now(),
	}
}

// Update updates the scheduler state
func (s *Scheduler) Update(dischargeInfo *monitors.DischargeInfo, isEmergency, acOnline bool) {
	oldMode := s.currentMode

	switch {
	case isEmergency:
		s.currentMode = ModeEmergency
	case dischargeInfo.IsFast:
		s.consecutiveFastDischarge++
		if s.consecutiveFastDischarge >= 2 {
			s.currentMode = ModeAggressive
		}
	default:
		s.consecutiveFastDischarge = 0
		s.currentMode = ModeNormal
	}

	if acOnline {
		s.currentMode = ModeNormal
		s.consecutiveFastDischarge = 0
	}

	if oldMode != s.currentMode {
		log.Printf("Mode change: %s -> %s", oldMode, s.currentMode)
	}
}

// ShouldCheckpoint returns true if it's time to checkpoint
func (s *Scheduler) ShouldCheckpoint() bool {
	interval := s.CurrentInterval()
	return time.Since(s.lastCheckpoint) >= interval
}

// RecordCheckpoint records that a checkpoint was saved
func (s *Scheduler) RecordCheckpoint() {
	s.lastCheckpoint = time.Now()
}

// CurrentInterval returns the current checkpoint interval
func (s *Scheduler) CurrentInterval() time.Duration {
	switch s.currentMode {
	case ModeAggressive:
		return s.aggressiveInterval
	case ModeEmergency:
		return s.emergencyInterval
	default:
		return s.normalInterval
	}
}

// CurrentMode returns the current mode
func (s *Scheduler) CurrentMode() Mode {
	return s.currentMode
}
