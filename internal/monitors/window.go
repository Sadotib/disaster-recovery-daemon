// internal/monitors/window.go
package monitors

import (
	"os/exec"
	"strings"
)

// ActiveWindow represents an X11 window
type ActiveWindow struct {
	Title string `json:"title"`
	Class string `json:"class"`
	PID   int    `json:"pid"`
}

// WindowMonitor monitors X11 windows
type WindowMonitor struct{}

// NewWindowMonitor creates a new window monitor
func NewWindowMonitor() *WindowMonitor {
	return &WindowMonitor{}
}

// GetActiveWindows returns list of visible windows
func (wm *WindowMonitor) GetActiveWindows() ([]ActiveWindow, error) {
	windows := make([]ActiveWindow, 0)

	// Check if xdotool is available
	if _, err := exec.LookPath("xdotool"); err != nil {
		return windows, nil
	}

	// Get list of windows
	out, err := exec.Command("xdotool", "search", "--onlyvisible", ".*").Output()
	if err != nil {
		return windows, nil
	}

	for _, windowID := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if windowID == "" {
			continue
		}

		title, _ := exec.Command("xdotool", "getwindowname", windowID).Output()
		class, _ := exec.Command("xdotool", "getwindowclassname", windowID).Output()

		windows = append(windows, ActiveWindow{
			Title: strings.TrimSpace(string(title)),
			Class: strings.TrimSpace(string(class)),
			PID:   0, // xdotool getwindowpid may fail
		})
	}

	if len(windows) > 20 {
		windows = windows[:20]
	}

	return windows, nil
}
