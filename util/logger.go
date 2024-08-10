package util

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func SetupLogger(logPath string) (*log.Logger, error) {
	projectRoot, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %v", err)
	}

	// Construct the absolute path for the log file
	absoluteLogPath := filepath.Join(projectRoot, logPath)

	// Ensure the directory exists
	dir := filepath.Dir(absoluteLogPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open the log file (create if not exists, append if exists)
	file, err := os.OpenFile(absoluteLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// Create a new logger
	return log.New(file, "", log.LstdFlags), nil
}
