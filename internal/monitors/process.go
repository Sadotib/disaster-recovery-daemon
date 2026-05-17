// internal/monitors/process.go
package monitors

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Process represents a running process
type Process struct {
	PID     int    `json:"pid"`
	Name    string `json:"name"`
	Command string `json:"command"`
	CWD     string `json:"cwd"`
}

// ProcessMonitor monitors running processes
type ProcessMonitor struct{}

// NewProcessMonitor creates a new process monitor
func NewProcessMonitor() *ProcessMonitor {
	return &ProcessMonitor{}
}

// GetProcesses returns list of running processes
func (pm *ProcessMonitor) GetProcesses() ([]Process, error) {
	processes := make([]Process, 0)

	entries, err := os.ReadDir("/proc")
	if err != nil {
		return processes, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		var pid int
		if _, err := fmt.Sscanf(entry.Name(), "%d", &pid); err != nil {
			continue
		}

		// Read process name
		commPath := filepath.Join("/proc", entry.Name(), "comm")
		name := ""
		if data, err := os.ReadFile(commPath); err == nil {
			name = strings.TrimSpace(string(data))
		}

		// Skip kernel processes
		if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
			continue
		}

		// Read command line
		cmdlinePath := filepath.Join("/proc", entry.Name(), "cmdline")
		command := name
		if data, err := os.ReadFile(cmdlinePath); err == nil && len(data) > 0 {
			cmd := strings.ReplaceAll(string(data), "\x00", " ")
			if len(cmd) > 0 {
				command = cmd
			}
		}

		// Read working directory
		cwdPath := filepath.Join("/proc", entry.Name(), "cwd")
		cwd, _ := os.Readlink(cwdPath)

		if name != "" && name != "power-recovery" && name != "idle" {
			processes = append(processes, Process{
				PID:     pid,
				Name:    name,
				Command: command,
				CWD:     cwd,
			})
		}
	}

	sort.Slice(processes, func(i, j int) bool {
		return processes[i].PID < processes[j].PID
	})

	if len(processes) > 100 {
		processes = processes[:100]
	}

	return processes, nil
}
