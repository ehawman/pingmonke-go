package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Color codes for terminal output
const (
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorMagenta = "\033[35m"
	ColorReset   = "\033[0m"
	ColorBold    = "\033[1m"
)

// PingLine represents a single line in the CSV with formatting
type PingLine struct {
	StartTime string
	EndTime   string
	Latency   int64
	Status    string
	Raw       []string
}

// GetColoredLine returns the ping line with appropriate color formatting
func (p *PingLine) GetColoredLine(widths []int) string {
	color := ColorReset
	switch {
	case p.Status == "timeout":
		color = ColorRed
	case p.Latency >= 100:
		color = ColorYellow
	case p.Latency > 0 && p.Latency < 100:
		color = ColorGreen
	default:
		color = ColorMagenta
	}

	// Format with padding
	line := fmt.Sprintf("%s%-*s %-*s %-*d %s%s",
		color,
		widths[0], p.StartTime,
		widths[1], p.EndTime,
		widths[2], p.Latency,
		p.Status,
		ColorReset)
	return line
}

// ReadPingFile reads the ping CSV file and returns records
func ReadPingFile(filePath string) ([]PingLine, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("empty file")
	}

	// Skip header
	var lines []PingLine
	for i := 1; i < len(records); i++ {
		if len(records[i]) < 4 {
			continue
		}

		latency, _ := strconv.ParseInt(records[i][2], 10, 64)
		lines = append(lines, PingLine{
			StartTime: records[i][0],
			EndTime:   records[i][1],
			Latency:   latency,
			Status:    records[i][3],
			Raw:       records[i],
		})
	}

	return lines, nil
}

// GetSummaryStats reads a ping file and returns summary statistics
func GetSummaryStats(filePath string) (total, ok, delayed, timeout int, avgLatency float64) {
	lines, err := ReadPingFile(filePath)
	if err != nil {
		return
	}

	total = len(lines)
	var totalLatency int64

	for _, line := range lines {
		switch line.Status {
		case "ok":
			ok++
		case "delayed":
			delayed++
		case "timeout":
			timeout++
		}
		if line.Status != "timeout" {
			totalLatency += line.Latency
		}
	}

	if ok+delayed > 0 {
		avgLatency = float64(totalLatency) / float64(ok+delayed)
	}

	return
}

// FormatHeader returns a formatted header line for the ping table
func FormatHeader() string {
	return fmt.Sprintf("%s%-27s %-27s %-6s %s%s",
		ColorBold,
		"Ping Init",
		"Ping Rec",
		"Time",
		"Status",
		ColorReset)
}

// GetColumnWidths calculates appropriate column widths for display
func GetColumnWidths(lines []PingLine) [3]int {
	widths := [3]int{27, 27, 6}
	for _, line := range lines {
		if len(line.StartTime) > widths[0] {
			widths[0] = len(line.StartTime)
		}
		if len(line.EndTime) > widths[1] {
			widths[1] = len(line.EndTime)
		}
		latencyLen := len(fmt.Sprintf("%d", line.Latency))
		if latencyLen > widths[2] {
			widths[2] = latencyLen
		}
	}
	return widths
}

// TruncateTime simplifies timestamp to YYYY-MM-DD HH:MM:SS.mmm format
func TruncateTime(timeStr string) string {
	// Parse RFC3339 format and return simplified version
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		// Try the new format too
		t, err = time.Parse("2006-01-02 15:04:05.000", timeStr)
		if err != nil {
			return timeStr
		}
	}
	return t.Format("2006-01-02 15:04:05.000")
}

// FormatSummaryLine returns a formatted summary statistics line
func FormatSummaryLine(total, ok, delayed, timeout int, avgLatency float64) string {
	return fmt.Sprintf("Total: %d | OK: %d | Delayed: %d | Timeout: %d | Avg: %.0fms",
		total, ok, delayed, timeout, avgLatency)
}
