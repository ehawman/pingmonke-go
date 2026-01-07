// cmd/tailmonke/main.go
// @version skeleton
// @description Entry point for tailmonke utility

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ehawman/pingmonke-go/internal"
)

func main() {
	file := flag.String("file", "", "Log file to tail")
	follow := flag.Bool("follow", false, "Follow mode (like tail -f)")
	summary := flag.Bool("summary", false, "Generate summary for file")
	flag.Parse()

	if *file == "" {
		fmt.Println("Usage: tailmonke --file <logfile> [--follow] [--summary]")
		os.Exit(1)
	}

	if *follow {
		internal.TailFollow(*file)
	} else {
		internal.PrintLastEntries(*file, 120)
	}

	if *summary {
		internal.GenerateSummary(*file)
	}
}
