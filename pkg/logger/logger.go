// pkg/logger/logger.go
package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	// LogFile is the file handle for logging
	LogFile *os.File
)

// Init sets up logging to both file and stdout
func Init(dataDir string, debug bool) error {
	// Create logs directory
	logDir := filepath.Join(dataDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	// Open log file
	logPath := filepath.Join(logDir, "daemon.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	LogFile = file

	// Set log output to both file and stdout
	if debug {
		log.SetOutput(io.MultiWriter(file, os.Stdout))
	} else {
		log.SetOutput(file)
	}

	return nil
}

// Close closes the log file
func Close() {
	if LogFile != nil {
		LogFile.Close()
	}
}
