package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Color codes for terminal output
const (
	ColorGreen   = "\033[32m"
	ColorRed     = "\033[31m"
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

	// Format with padding, adding "ms" suffix to latency
	line := fmt.Sprintf("%s%-*s %-*s %-*s %-12s%s",
		color,
		widths[0], p.StartTime,
		widths[1], p.EndTime,
		widths[2], fmt.Sprintf("%dms", p.Latency),
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

// FormatHeader returns a formatted header line for the ping table with health color
func FormatHeader(healthColor string) string {
	// Map health colors to background colors
	var bgColor string
	switch healthColor {
	case ColorRed:
		bgColor = "160" // Dark red
	case ColorYellow:
		bgColor = "136" // Dark yellow/orange
	default:
		bgColor = "22" // Dark green
	}

	// Create style with health color background (no padding to preserve alignment)
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("white")).
		Background(lipgloss.Color(bgColor))

	cols := []string{
		"Ping Init",
		"Ping Rec",
		"Time",
		"Status",
	}

	// Format header: use the styled columns but without padding
	// Match the exact widths used in GetColoredLine
	header := fmt.Sprintf("%-27s %-27s %-8s %-12s",
		cols[0],
		cols[1],
		cols[2],
		cols[3])

	return headerStyle.Render(header)
}

// GetColumnWidths calculates appropriate column widths for display
func GetColumnWidths(lines []PingLine) [3]int {
	widths := [3]int{27, 27, 8}
	for _, line := range lines {
		if len(line.StartTime) > widths[0] {
			widths[0] = len(line.StartTime)
		}
		if len(line.EndTime) > widths[1] {
			widths[1] = len(line.EndTime)
		}
		// Account for "ms" suffix in latency
		latencyLen := len(fmt.Sprintf("%dms", line.Latency))
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

// FormatSummaryLine returns a formatted summary statistics line with color coding
func FormatSummaryLine(total, ok, delayed, timeout int, avgLatency float64) string {
	// Determine most severe status for Total color
	var totalColor string
	switch {
	case timeout > 0:
		totalColor = ColorRed
	case delayed > 0:
		totalColor = ColorYellow
	default:
		totalColor = ColorGreen
	}

	// Determine avg color
	var avgColor string
	switch {
	case avgLatency >= 100:
		avgColor = ColorYellow
	case avgLatency > 0:
		avgColor = ColorGreen
	default:
		avgColor = ColorMagenta
	}

	return fmt.Sprintf("%sTotal: %d%s | %sOK: %d%s | %sDelayed: %d%s | %sTimeout: %d%s | %sAvg: %.0fms%s",
		totalColor, total, ColorReset,
		ColorGreen, ok, ColorReset,
		ColorYellow, delayed, ColorReset,
		ColorRed, timeout, ColorReset,
		avgColor, avgLatency, ColorReset)
}

// FormatEventLine returns a formatted event status line
func FormatEventLine(event EventStatus, healthColor string) string {
	var statusCircle string
	var statusColor string

	switch healthColor {
	case ColorRed:
		statusCircle = "●"
		statusColor = ColorRed
	case ColorYellow:
		statusCircle = "◐"
		statusColor = ColorYellow
	default:
		statusCircle = "○"
		statusColor = ColorGreen
	}

	// Determine Last/Current based on whether there's an active event
	var durationStr string
	if event.IsActive {
		// Event is active - show current duration
		durationStr = fmt.Sprintf("Current Event Duration: %v", event.Duration.Round(time.Second))
	} else if !event.EndTime.IsZero() {
		// Event has ended - show the ended duration
		durationStr = fmt.Sprintf("Last Event Duration: %v", event.Duration.Round(time.Second))
	} else {
		durationStr = "Last Event Duration: N/A"
	}

	return fmt.Sprintf("Event: %s%s%s | %s%s%s",
		statusColor, statusCircle, ColorReset,
		statusColor, durationStr, ColorReset)
}
