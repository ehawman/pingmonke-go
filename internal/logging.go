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

	// Create main CSV file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("[Logging] Error creating log file:", err)
		return filename
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Write([]string{"Ping Init", "Ping Rec", "Ping Time (ms)", "Status"})
	writer.Flush()

	// Create test format file
	testFilename := filename[:len(filename)-4] + "-test.csv"
	testFile, err := os.Create(testFilename)
	if err != nil {
		fmt.Println("[Logging] Error creating test log file:", err)
	} else {
		defer testFile.Close()
		testWriter := csv.NewWriter(testFile)
		testWriter.Write([]string{"Response"})
		testWriter.Flush()
	}

	fmt.Println("[Logging] Log file ready:", filename)
	return filename
}

// writeToCSV appends a ping result to both main and test log files.
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

	// Write to test CSV
	testFile := file[:len(file)-4] + "-test.csv"
	tf, err := os.OpenFile(testFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// Test file might not exist, skip
		return
	}
	defer tf.Close()

	testWriter := csv.NewWriter(tf)
	testRow := []string{formatTestLine(latency, status)}
	testWriter.Write(testRow)
	testWriter.Flush()
}

// formatTimestamp converts a time to the standardized format: YYYY-MM-DD HH:MM:SS.mmm
func formatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000")
}

// formatTestLine formats a ping result in ping command output format
func formatTestLine(latency time.Duration, status string) string {
	if status == "timeout" {
		return "ping: unknown host"
	}
	ms := float64(latency.Milliseconds())
	if ms == 0 && status == "ok" {
		ms = 0.1 // Avoid 0ms for successful pings
	}
	return fmt.Sprintf("64 bytes from localhost: icmp_seq=1 ttl=104 time=%.1f ms", ms)
}
