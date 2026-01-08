// internal/logging.go
// @version improved
// @description Handles CSV logging for ping results with directory fallback.

package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"
)

// prepareLogFile creates a new log file for the current period.
func prepareLogFile(periodStart time.Time, cfg Config) string {
	logDir := cfg.LogDir
	if logDir == "" {
		logDir = defaultLogDir() // fallback if env/config missing
	}
	logDir = ExpandHome(logDir) // expand ~ to home directory
	var filename string
	if cfg.DebugMode {
		filename = fmt.Sprintf("%s/%s-pings.csv", logDir, periodStart.Format("15:04:05.000"))
	} else {
		filename = fmt.Sprintf("%s/%s-pings.csv", logDir, periodStart.Format("2006-01-02"))
	}

	// Ensure directory exists
	err := os.MkdirAll(logDir, 0755)
	if err != nil {
		fmt.Println("[Logging] Error creating log directory:", err)
	}

	// Check if file already exists and has content (header)
	fileInfo, err := os.Stat(filename)
	if err == nil && fileInfo.Size() > 0 {
		// File exists and has content, don't recreate it
		fmt.Println("[Logging] Log file ready:", filename)
		return filename
	}

	// File doesn't exist or is empty, create it with header
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("[Logging] Error creating log file:", err)
		return filename
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Write([]string{"Ping Init", "Ping Rec", "Ping Time (ms)", "Status"})
	writer.Flush()

	fmt.Println("[Logging] Log file ready:", filename)
	return filename
}

// writeToCSV appends a ping result to the main log file.
func writeToCSV(file string, start, end time.Time, latency time.Duration, status string) {
	// Write to main CSV
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("[Logging] Error opening log file:", err)
		return
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	row := []string{
		formatTimestamp(start),
		formatTimestamp(end),
		fmt.Sprintf("%d", latency.Milliseconds()),
		status,
	}
	writer.Write(row)
	writer.Flush()
}

// formatTimestamp converts a time to the standardized format: YYYY-MM-DD HH:MM:SS.mmm
func formatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000")
}
