// internal/checkpoint/manager.go
package checkpoint

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// Manager handles checkpoint operations
type Manager struct {
	dataDir        string
	maxCheckpoints int
	dryRun         bool
}

// NewManager creates a new checkpoint manager
func NewManager(dataDir string, maxCheckpoints int, dryRun bool) *Manager {
	return &Manager{
		dataDir:        dataDir,
		maxCheckpoints: maxCheckpoints,
		dryRun:         dryRun,
	}
}

// Save saves a checkpoint
func (m *Manager) Save(checkpoint *Checkpoint) error {
	if m.dryRun {
		log.Printf("[DRY RUN] Would save checkpoint: %s", checkpoint.SaveReason)
		return nil
	}

	timestamp := checkpoint.Timestamp.Format("20060102_150405")
	checkpointPath := filepath.Join(m.dataDir, fmt.Sprintf("checkpoint_%s.json", timestamp))
	latestPath := filepath.Join(m.dataDir, "latest_checkpoint.json")
	tmpPath := filepath.Join(m.dataDir, "checkpoint.tmp")

	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}

	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp: %w", err)
	}

	if err := os.Rename(tmpPath, checkpointPath); err != nil {
		return fmt.Errorf("failed to rename: %w", err)
	}

	// Also update latest
	os.Rename(tmpPath, latestPath)

	m.cleanup()

	log.Printf("Checkpoint saved: %s (%.1f%% battery)", checkpoint.SaveReason, checkpoint.BatteryPercent)
	return nil
}

// LoadLatest loads the most recent checkpoint
func (m *Manager) LoadLatest() (*Checkpoint, error) {
	latestPath := filepath.Join(m.dataDir, "latest_checkpoint.json")

	data, err := os.ReadFile(latestPath)
	if err != nil {
		return nil, err
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, err
	}

	return &checkpoint, nil
}

func (m *Manager) cleanup() {
	files, _ := filepath.Glob(filepath.Join(m.dataDir, "checkpoint_*.json"))

	sort.Slice(files, func(i, j int) bool {
		infoI, _ := os.Stat(files[i])
		infoJ, _ := os.Stat(files[j])
		if infoI == nil || infoJ == nil {
			return false
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	for len(files) >= m.maxCheckpoints {
		os.Remove(files[0])
		files = files[1:]
	}
}
