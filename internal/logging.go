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

// writeToCSV appends a ping result to the log file.
func writeToCSV(file string, start, end time.Time, latency time.Duration, status string) {
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("[Logging] Error opening log file:", err)
		return
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	row := []string{
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
		fmt.Sprintf("%d", latency.Milliseconds()),
		status,
	}
	writer.Write(row)
	writer.Flush()
}
