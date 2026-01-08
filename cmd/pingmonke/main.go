// cmd/pingmonke/main.go
// @version skeleton
// @description Entry point for pingmonke service

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ehawman/pingmonke-go/internal"
)

func main() {
	verbose := flag.Bool("v", false, "Enable verbose mode")
	debug := flag.Bool("debug-rollover", false, "Enable debug rollover mode")
	configPath := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	cfg := internal.LoadConfig(*configPath)
	cfg.Verbose = *verbose
	cfg.DebugMode = *debug

	// Auto-enable verbose mode when in debug mode
	if cfg.DebugMode {
		cfg.Verbose = true
	}

	internal.SetDefaults(&cfg)
	internal.PrepareLogDirectory(cfg.LogDir)

	fmt.Println("Starting pingmonke service...")
	internal.StartScheduler(cfg)

	os.Exit(0)
}
