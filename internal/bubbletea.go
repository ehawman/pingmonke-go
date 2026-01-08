package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TailmonkeModel represents the state of the TUI
type TailmonkeModel struct {
	filePath         string
	logDir           string
	linesToDisplay   int
	lines            []PingLine
	width            int
	height           int
	lastError        string
	lastRefresh      time.Time
	summaryLine      string
	columnWidths     [3]int
	newFilePath      string      // Detected new file available for switching
	lastFileCheck    time.Time   // Last time we checked for new files
	explicitFile     bool        // If true, user provided --file flag, disable file detection
	lastNotification string      // Last notification message to display
	notificationTime time.Time   // When the notification was set
	eventStatus      EventStatus // Current event status
}

// Message types for bubbletea
type FileUpdatedMsg struct {
	lines []PingLine
	time  time.Time
}

type TickMsg time.Time

// NewTailmonkeModel creates a new TUI model
func NewTailmonkeModel(filePath string, linesToDisplay int) *TailmonkeModel {
	return NewTailmonkeModelWithOptions(filePath, linesToDisplay, false)
}

// NewTailmonkeModelWithOptions creates a new TUI model with explicit options
func NewTailmonkeModelWithOptions(filePath string, linesToDisplay int, explicitFile bool) *TailmonkeModel {
	logDir := filepath.Dir(filePath)
	return &TailmonkeModel{
		filePath:       filePath,
		logDir:         logDir,
		linesToDisplay: linesToDisplay,
		lastRefresh:    time.Now(),
		lastFileCheck:  time.Now(),
		explicitFile:   explicitFile,
	}
}

// Init initializes the model and starts the tick timer
func (m *TailmonkeModel) Init() tea.Cmd {
	return tea.Batch(
		m.loadFile(),
		m.tickCmd(),
	)
}

// Update handles messages
func (m *TailmonkeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "f5", "ctrl+r":
			// Regenerate summary for current file
			m.regenerateSummary()
			return m, m.loadFile()
		case "n", "N":
			// Switch to new file if available
			if m.newFilePath != "" {
				m.filePath = m.newFilePath
				m.newFilePath = ""
				return m, m.loadFile()
			}
		}

	case FileUpdatedMsg:
		m.lines = msg.lines
		m.lastRefresh = msg.time
		m.updateSummary()
		m.updateHealthState() // Update health state whenever file updates
		return m, m.tickCmd()

	case TickMsg:
		// Check if file has been updated
		info, err := os.Stat(m.filePath)
		if err == nil && info.ModTime().After(m.lastRefresh.Add(-time.Second)) {
			return m, m.loadFile()
		}

		// Update health state on every tick (for duration display updates)
		m.updateHealthState()

		// Check for new files every 5 seconds (only if file was auto-detected)
		if !m.explicitFile && time.Since(m.lastFileCheck) >= 5*time.Second {
			m.lastFileCheck = time.Now()
			newFile := m.findNewerLogFile()
			if newFile != "" {
				m.newFilePath = newFile
			}
		}

		// Force re-render to update event duration display
		// This ensures the duration string updates every second
		return m, m.tickCmd()
	}

	return m, nil
}

// View renders the TUI
func (m *TailmonkeModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	output := ""

	// Determine health color based on event status
	var healthColor string
	if m.eventStatus.IsActive {
		healthColor = ColorRed
	} else if !m.eventStatus.EndTime.IsZero() {
		// Event has ended and recovered - show green
		healthColor = ColorGreen
	} else {
		healthColor = ColorGreen
	}

	// Calculate available space
	headerSpace := 1  // Header row
	notifySpace := 1  // Notification line (always present)
	summarySpace := 1 // Summary stats line
	eventSpace := 1   // Event status line
	availableHeight := m.height - headerSpace - notifySpace - summarySpace - eventSpace

	if availableHeight < 3 {
		return "Terminal too small"
	}

	// 1. Header - with health color background
	output += FormatHeader(healthColor) + "\n"

	// 2. Ping data lines - use all available space
	linesToShow := availableHeight
	if linesToShow < 1 {
		linesToShow = 1
	}

	// Calculate starting position
	startIdx := len(m.lines) - linesToShow
	if startIdx < 0 {
		startIdx = 0
	}

	// Apply dimming if new file available
	visibleLines := m.lines[startIdx:]
	for _, line := range visibleLines {
		lineStr := line.GetColoredLine(m.columnWidths[:])
		// Dim the line if new file is available
		if m.newFilePath != "" {
			lineStr = fmt.Sprintf("\033[2m%s\033[0m", lineStr) // Dim effect
		}
		output += lineStr + "\n"
	}

	// Pad remaining space
	for i := len(visibleLines); i < linesToShow; i++ {
		output += "\n"
	}

	// 3. Notification section
	// Priority: new file notification > summary notification
	if m.newFilePath != "" {
		newFileName := filepath.Base(m.newFilePath)
		notifyMsg := fmt.Sprintf("📢 New log file available: %s  Press 'N' to switch", newFileName)
		output += fmt.Sprintf("%s%s%s\n", ColorYellow, notifyMsg, ColorReset)
	} else if m.lastNotification != "" && time.Since(m.notificationTime) < 5*time.Second {
		// Show summary notifications (from F5 refresh, etc)
		notifMsg := strings.ReplaceAll(m.lastNotification, "\n", " ")
		if len(notifMsg) > m.width-1 {
			notifMsg = notifMsg[:m.width-4] + "..."
		}
		output += fmt.Sprintf("%s%s%s\n", ColorYellow, notifMsg, ColorReset)
	} else {
		output += "\n" // Empty notification line
	}

	// 4. Summary statistics line
	output += m.summaryLine + "\n"

	// 5. Event status line
	eventLine := FormatEventLine(m.eventStatus, healthColor)
	output += eventLine

	if m.lastError != "" {
		output += fmt.Sprintf("\n%sError: %s%s", ColorRed, m.lastError, ColorReset)
	}

	return output
}

// loadFile reads the ping file and returns a message
func (m *TailmonkeModel) loadFile() tea.Cmd {
	return func() tea.Msg {
		lines, err := ReadPingFile(m.filePath)
		if err != nil {
			m.lastError = err.Error()
			return FileUpdatedMsg{lines: []PingLine{}, time: time.Now()}
		}
		m.lastError = ""
		m.updateColumnWidths(lines)
		return FileUpdatedMsg{lines: lines, time: time.Now()}
	}
}

// updateColumnWidths calculates column widths for the current data
func (m *TailmonkeModel) updateColumnWidths(lines []PingLine) {
	m.columnWidths = GetColumnWidths(lines)
}

// updateSummary calculates and formats the summary line
func (m *TailmonkeModel) updateSummary() {
	total, ok, delayed, timeout, avg := GetSummaryStats(m.filePath)
	m.summaryLine = FormatSummaryLine(total, ok, delayed, timeout, avg)
}

// updateHealthState updates the health color based on event status
func (m *TailmonkeModel) updateHealthState() {
	m.eventStatus = DetectEvent(m.lines)
}

// regenerateSummary regenerates the summary file for the current log
func (m *TailmonkeModel) regenerateSummary() {
	cfg := &Config{DebugMode: false}                          // Assume normal mode for summary generation
	msg := generateSummaryWithLogging(m.filePath, *cfg, true) // capture message
	m.lastNotification = msg
	m.notificationTime = time.Now()
}

// findNewerLogFile checks if there's a newer log file than the current one
func (m *TailmonkeModel) findNewerLogFile() string {
	entries, err := os.ReadDir(m.logDir)
	if err != nil {
		return ""
	}

	currentInfo, err := os.Stat(m.filePath)
	if err != nil {
		return ""
	}

	var newerFile string
	var newerTime time.Time

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Only consider main log files, not test files
		if filepath.Ext(name) == ".csv" && len(name) > 10 && name[len(name)-10:] == "-pings.csv" {
			info, err := entry.Info()
			if err != nil {
				continue
			}

			// Check if this file is newer than current and newer than what we've already found
			if info.ModTime().After(currentInfo.ModTime()) && info.ModTime().After(newerTime) {
				fullPath := filepath.Join(m.logDir, name)
				// Don't suggest switching to a file we're already viewing
				if fullPath != m.filePath {
					newerFile = fullPath
					newerTime = info.ModTime()
				}
			}
		}
	}

	return newerFile
}

// tickCmd returns a command that ticks every second to check for file updates
func (m *TailmonkeModel) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return TickMsg(time.Now())
	})
}

// RunTailmonkeTUI starts the interactive TUI (backward compatible)
func RunTailmonkeTUI(filePath string, linesToDisplay int) error {
	return RunTailmonkeTUIWithOptions(filePath, linesToDisplay, false)
}

// RunTailmonkeTUIWithOptions starts the interactive TUI with explicit file flag
func RunTailmonkeTUIWithOptions(filePath string, linesToDisplay int, explicitFile bool) error {
	model := NewTailmonkeModelWithOptions(filePath, linesToDisplay, explicitFile)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
