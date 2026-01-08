// cmd/tailmonke/main.go
// @version tui
// @description TUI for tailing ping logs with real-time stats and syntax highlighting

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ehawman/pingmonke-go/internal"
)

func main() {
	file := flag.String("file", "", "Log file to tail")
	configPath := flag.String("config", "config.yaml", "Path to config file")
	nonInteractive := flag.Bool("non-interactive", false, "Non-interactive mode (plain output)")
	flag.Parse()

	// Load config
	cfg := internal.LoadConfig(*configPath)

	// Track if file was explicitly provided
	explicitFile := *file != ""

	// If no file specified, try to find the most recent log file from config
	if *file == "" {
		*file = findMostRecentLogFile(internal.ExpandHome(cfg.LogDir))
		if *file == "" {
			fmt.Println("Error: No ping log file found")
			os.Exit(1)
		}
	}

	if *nonInteractive {
		// Simple non-interactive output
		displayNonInteractive(*file, cfg.Tailmonke.LinesToDisplay)
	} else {
		// Interactive TUI mode with bubbletea
		err := internal.RunTailmonkeTUIWithOptions(*file, cfg.Tailmonke.LinesToDisplay, explicitFile)
		if err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
			os.Exit(1)
		}
	}
}

// displayNonInteractive shows the log file in simple text mode
func displayNonInteractive(filePath string, linesToDisplay int) {
	lines, err := internal.ReadPingFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Display header with green (default) health color
	fmt.Println(internal.FormatHeader(internal.ColorGreen))
	fmt.Println(strings.Repeat("-", 70))

	// Show last N lines
	start := len(lines) - linesToDisplay
	if start < 0 {
		start = 0
	}

	widths := internal.GetColumnWidths(lines[start:])
	for _, line := range lines[start:] {
		fmt.Println(line.GetColoredLine(widths[:]))
	}

	// Show summary
	fmt.Println(strings.Repeat("-", 70))
	total, ok, delayed, timeout, avg := internal.GetSummaryStats(filePath)
	fmt.Println(internal.FormatSummaryLine(total, ok, delayed, timeout, avg))
}

// findMostRecentLogFile finds the most recently modified *-pings.csv file in the log directory
func findMostRecentLogFile(logDir string) string {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		fmt.Printf("[Error] Could not read log directory %s: %v\n", logDir, err)
		return ""
	}

	var logFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".csv" {
			name := entry.Name()
			// Check if it matches the pattern *-pings.csv
			if len(name) > 10 && name[len(name)-10:] == "-pings.csv" {
				logFiles = append(logFiles, entry)
			}
		}
	}

	if len(logFiles) == 0 {
		return ""
	}

	// Sort by modification time, newest first
	sort.Slice(logFiles, func(i, j int) bool {
		iInfo, _ := logFiles[i].Info()
		jInfo, _ := logFiles[j].Info()
		return iInfo.ModTime().After(jInfo.ModTime())
	})

	return filepath.Join(logDir, logFiles[0].Name())
}
