// internal/common.go
// @version improved
// @description Shared helpers and directory setup for pingmonke and tailmonke

package internal

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds global settings for pingmonke.
type Config struct {
	Target        string          `yaml:"target"`
	LogDir        string          `yaml:"log_dir"`
	Interval      time.Duration   `yaml:"interval"`
	DebugInterval time.Duration   `yaml:"debug_interval"`
	Port          int             `yaml:"port"`
	UseICMP       bool            `yaml:"use_icmp"`
	Tailmonke     TailmonkeConfig `yaml:"tailmonke"`
	Verbose       bool
	DebugMode     bool
}

// TailmonkeConfig holds tailmonke-specific settings
type TailmonkeConfig struct {
	LinesToDisplay int `yaml:"lines_to_display"`
}

// LoadConfig reads and parses the YAML config file.
func LoadConfig(path string) Config {
	cfg := Config{
		Target:        "google.com",
		LogDir:        defaultLogDir(),
		Interval:      15 * time.Second,
		DebugInterval: 5 * time.Second,
		Port:          80,
		UseICMP:       false,
		Tailmonke: TailmonkeConfig{
			LinesToDisplay: 20,
		},
	}

	// Try to read the config file
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, use defaults
			fmt.Printf("[Config] Config file not found at %s, using defaults\n", path)
		} else {
			fmt.Printf("[Config] Error reading config file: %v\n", err)
		}
		return cfg
	}

	// Parse YAML
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		fmt.Printf("[Config] Error parsing YAML: %v, using defaults\n", err)
		return cfg
	}

	fmt.Printf("[Config] Loaded config from %s\n", path)
	return cfg
}

// SetDefaults placeholder.
func SetDefaults(cfg *Config) {
	if cfg.LogDir == "" {
		cfg.LogDir = defaultLogDir()
	}
}

// PrepareLogDirectory ensures the log directory exists.
func PrepareLogDirectory(dir string) {
	dir = ExpandHome(dir) // expand ~ to home directory
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

// ExpandHome expands ~ to the user's home directory
func ExpandHome(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if len(path) == 1 {
			return home
		}
		if path[1] == '/' {
			return home + path[1:]
		}
	}
	return path
}

// Tailmonke stubs:
func TailFollow(file string)              {}
func PrintLastEntries(file string, n int) {}
func GenerateSummary(file string)         {}
