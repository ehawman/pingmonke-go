package internal

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"
)

// PingRecord represents a single ping entry
type PingRecord struct {
	StartTime time.Time
	EndTime   time.Time
	Latency   int64 // milliseconds
	Status    string
}

// Event represents a network event (outage/degradation)
type Event struct {
	StartIndex int // index in records
	EndIndex   int // index in records
	StartTime  time.Time
	EndTime    time.Time
}

func generateSummary(logFile string, cfg Config) {
	f, err := os.Open(logFile)
	if err != nil {
		fmt.Println("[Summary] Error opening log file:", err)
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, _ := reader.ReadAll()
	records = records[1:] // skip header

	// Parse records
	var pings []PingRecord
	okCount, delayedCount, timeoutCount := 0, 0, 0
	for _, row := range records {
		startTime, _ := time.Parse(time.RFC3339, row[0])
		endTime, _ := time.Parse(time.RFC3339, row[1])
		latency, _ := strconv.ParseInt(row[2], 10, 64)
		status := row[3]

		pings = append(pings, PingRecord{
			StartTime: startTime,
			EndTime:   endTime,
			Latency:   latency,
			Status:    status,
		})

		switch status {
		case "ok":
			okCount++
		case "delayed":
			delayedCount++
		case "timeout":
			timeoutCount++
		}
	}

	// Detect events
	events := detectEvents(pings, cfg.DebugMode)

	summaryFile := logFile[:len(logFile)-4] + "-summary.csv"
	sf, err := os.Create(summaryFile)
	if err != nil {
		fmt.Println("[Summary] Error creating summary file:", err)
		return
	}
	defer sf.Close()

	writer := csv.NewWriter(sf)

	// Write summary counts
	writer.Write([]string{"Total", "OK", "Delayed", "Timeout", "Events"})
	writer.Write([]string{
		fmt.Sprintf("%d", len(pings)),
		fmt.Sprintf("%d", okCount),
		fmt.Sprintf("%d", delayedCount),
		fmt.Sprintf("%d", timeoutCount),
		fmt.Sprintf("%d", len(events)),
	})
	writer.Write([]string{}) // blank line

	// Write events
	if len(events) > 0 {
		for i, event := range events {
			writer.Write([]string{fmt.Sprintf("Event %d", i+1)})
			writer.Write([]string{
				"Start",
				formatTimestampForSummary(event.StartTime),
			})
			writer.Write([]string{
				"End",
				formatTimestampForSummary(event.EndTime),
			})
			duration := event.EndTime.Sub(event.StartTime)
			writer.Write([]string{
				"Duration",
				fmt.Sprintf("%v", duration),
			})

			// Get context: 3 pings before and after
			startContext := event.StartIndex - 3
			if startContext < 0 {
				startContext = 0
			}
			endContext := event.EndIndex + 3
			if endContext >= len(pings) {
				endContext = len(pings) - 1
			}

			writer.Write([]string{}) // blank line
			writer.Write([]string{"Ping Init", "Ping Rec", "Ping Time (ms)", "Status"})

			for j := startContext; j <= endContext; j++ {
				p := pings[j]
				prefix := ""
				if j < event.StartIndex || j > event.EndIndex {
					prefix = "* " // mark context pings
				}
				writer.Write([]string{
					prefix + formatTimestampForSummary(p.StartTime),
					formatTimestampForSummary(p.EndTime),
					fmt.Sprintf("%d", p.Latency),
					p.Status,
				})
			}
			writer.Write([]string{}) // blank line
		}
	}

	writer.Flush()
	fmt.Printf("[Summary] Summary written to %s\n", summaryFile)

	// Also create a test summary file
	testSummaryFile := logFile[:len(logFile)-4] + "-test-summary.csv"
	tsf, err := os.Create(testSummaryFile)
	if err != nil {
		fmt.Println("[Summary] Error creating test summary file:", err)
		return
	}
	defer tsf.Close()

	testWriter := csv.NewWriter(tsf)
	testWriter.Write([]string{"Total", "OK", "Delayed", "Timeout", "Events"})
	testWriter.Write([]string{
		fmt.Sprintf("%d", len(pings)),
		fmt.Sprintf("%d", okCount),
		fmt.Sprintf("%d", delayedCount),
		fmt.Sprintf("%d", timeoutCount),
		fmt.Sprintf("%d", len(events)),
	})
	testWriter.Flush()
	fmt.Printf("[Summary] Test summary written to %s\n", testSummaryFile)
}

// formatTimestampForSummary formats a timestamp as YYYY-MM-DD HH:MM:SS.mmm
func formatTimestampForSummary(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000")
}

// detectEvents finds network events based on ping quality
// For debug mode: event starts with 2+ bad pings within 15 seconds, ends after 15 seconds of good pings
// For normal mode: event starts with 2+ bad pings within 60 seconds, ends after 60 seconds of good pings
func detectEvents(pings []PingRecord, debugMode bool) []Event {
	var events []Event

	windowDuration := 60 * time.Second
	if debugMode {
		windowDuration = 15 * time.Second
	}

	i := 0
	for i < len(pings) {
		// Look for event start: 2 bad pings within window duration
		eventStart := -1
		for j := i; j < len(pings)-1; j++ {
			if isBadPing(pings[j]) && isBadPing(pings[j+1]) {
				// Check if they're within window duration
				timeDiff := pings[j+1].StartTime.Sub(pings[j].StartTime)
				if timeDiff <= windowDuration {
					eventStart = j
					break
				}
			}
		}

		if eventStart == -1 {
			break
		}

		// Find event end: windowDuration of good pings
		eventEnd := eventStart
		goodCount := 0
		goodStartTime := time.Time{}

		for j := eventStart; j < len(pings); j++ {
			if isBadPing(pings[j]) {
				goodCount = 0
				goodStartTime = time.Time{}
				eventEnd = j
			} else {
				if goodCount == 0 {
					goodStartTime = pings[j].StartTime
				}
				goodCount++
				if goodStartTime.Add(windowDuration).Before(pings[j].StartTime) || goodStartTime.Add(windowDuration).Equal(pings[j].StartTime) {
					// We have windowDuration of good pings
					eventEnd = j
					events = append(events, Event{
						StartIndex: eventStart,
						EndIndex:   eventEnd,
						StartTime:  pings[eventStart].StartTime,
						EndTime:    pings[eventEnd].EndTime,
					})
					i = j + 1
					goto nextEvent
				}
			}
		}

		// Event goes to end of period
		if eventStart < len(pings) {
			events = append(events, Event{
				StartIndex: eventStart,
				EndIndex:   len(pings) - 1,
				StartTime:  pings[eventStart].StartTime,
				EndTime:    pings[len(pings)-1].EndTime,
			})
		}
		break

	nextEvent:
	}

	return events
}

func isBadPing(p PingRecord) bool {
	return p.Status == "timeout" || p.Latency >= 100
}
