// internal/common.go
// @version improved
// @description Shared helpers and directory setup for pingmonke and tailmonke

package internal

import (
	"fmt"
	"os"
	"time"
)

// Config holds global settings for pingmonke.
type Config struct {
	Target    string
	LogDir    string
	Interval  time.Duration
	Port      int
	UseICMP   bool
	Verbose   bool
	DebugMode bool
}

// LoadConfig placeholder (currently ignores YAML).
func LoadConfig(path string) Config {
	return Config{
		Target:   "google.com",
		LogDir:   defaultLogDir(),
		Interval: 15 * time.Second,
		Port:     80,
		UseICMP:  false,
	}
}

// SetDefaults placeholder.
func SetDefaults(cfg *Config) {
	if cfg.LogDir == "" {
		cfg.LogDir = defaultLogDir()
	}
}

// PrepareLogDirectory ensures the log directory exists.
func PrepareLogDirectory(dir string) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Println("[Setup] Error creating log directory:", err)
	} else {
		fmt.Println("[Setup] Log directory ready:", dir)
	}
}

// defaultLogDir returns $HOME/ping-logs if nothing else is set.
func defaultLogDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/ping-logs"
	}
	return fmt.Sprintf("%s/ping-logs", home)
}

// Concurrency stub for scheduler.
func waitForAllPings() {}

// Tailmonke stubs:
func TailFollow(file string)              {}
func PrintLastEntries(file string, n int) {}
func GenerateSummary(file string)         {}
