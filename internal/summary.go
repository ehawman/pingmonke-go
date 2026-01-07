package internal

import (
	"encoding/csv"
	"fmt"
	"os"
)

func generateSummary(logFile string) {
	f, err := os.Open(logFile)
	if err != nil {
		fmt.Println("[Summary] Error opening log file:", err)
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, _ := reader.ReadAll()
	records = records[1:] // skip header

	okCount, delayedCount, timeoutCount := 0, 0, 0
	for _, row := range records {
		status := row[3]
		switch status {
		case "ok":
			okCount++
		case "delayed":
			delayedCount++
		case "timeout":
			timeoutCount++
		}
	}

	summaryFile := logFile[:len(logFile)-4] + "-summary.csv"
	sf, err := os.Create(summaryFile)
	if err != nil {
		fmt.Println("[Summary] Error creating summary file:", err)
		return
	}
	defer sf.Close()

	writer := csv.NewWriter(sf)
	writer.Write([]string{"Total", "OK", "Delayed", "Timeout"})
	writer.Write([]string{
		fmt.Sprintf("%d", len(records)),
		fmt.Sprintf("%d", okCount),
		fmt.Sprintf("%d", delayedCount),
		fmt.Sprintf("%d", timeoutCount),
	})
	writer.Flush()

	fmt.Printf("[Summary] Summary written to %s\n", summaryFile)
}
