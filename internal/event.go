package internal

import (
	"time"
)

// EventStatus represents the current state of a network event
type EventStatus struct {
	IsActive       bool      // True if event is currently ongoing
	StartTime      time.Time // Timestamp of the first bad ping in the event cluster
	EndTime        time.Time // Timestamp of the first of the 4 good pings that ended the event (zero if still active)
	LastBadPing    time.Time // Timestamp of the most recent bad ping
	LatestPingTime time.Time // Timestamp of the latest ping in the dataset
	Duration       time.Duration
}

// DetectEvent analyzes ping data to detect network events
// Algorithm:
// 1. Check last 4 pings for active event (2+ bad pings)
// 2. If no active event, walk backward to find the most recent event cluster
// 3. Find event start by continuing backward until 4 consecutive good pings (or start of data)
// 4. Scan forward from start to find when event ends (if it does)
// Duration:
//   - Active event: from start time to latest ping
//   - Ended event: from first bad ping to first of the 4 ending good pings
func DetectEvent(lines []PingLine) EventStatus {
	event := EventStatus{}

	if len(lines) == 0 {
		return event
	}

	// Set latest ping time
	latestTime, err := parsePingTime(lines[len(lines)-1].StartTime)
	if err == nil {
		event.LatestPingTime = latestTime
	}

	// Check the last 4 pings for an active event (only if we have at least 4 pings)
	if len(lines) >= 4 {
		startIdx := len(lines) - 4
		lastFourPings := lines[startIdx:]

		// Count bad pings in last 4
		var badPingsInLastFour []PingLine
		for _, line := range lastFourPings {
			if line.Status == "timeout" || line.Status == "delayed" {
				badPingsInLastFour = append(badPingsInLastFour, line)
			}
		}

		// Rule 1: Check if event is active (2+ bad pings in last 4)
		if len(badPingsInLastFour) >= 2 {
			event.IsActive = true

			// Event start = oldest of the 2+ bad pings in last 4
			var times []time.Time
			for _, line := range badPingsInLastFour {
				t, err := parsePingTime(line.StartTime)
				if err == nil {
					times = append(times, t)
				}
			}

			if len(times) >= 2 {
				oldest := times[0]
				for _, t := range times {
					if t.Before(oldest) {
						oldest = t
					}
				}
				event.StartTime = oldest
			}

			// Find most recent bad ping
			if len(badPingsInLastFour) > 0 {
				last, err := parsePingTime(badPingsInLastFour[len(badPingsInLastFour)-1].StartTime)
				if err == nil {
					event.LastBadPing = last
				}
			}

			// Calculate duration: start to latest ping
			if !event.StartTime.IsZero() && !event.LatestPingTime.IsZero() {
				event.Duration = event.LatestPingTime.Sub(event.StartTime)
			}

			return event
		}
	}

	// No active event in last 4 pings - look for the most recent event by walking backward
	eventStartIdx := findMostRecentEventStart(lines)
	if eventStartIdx == -1 {
		// No previous event found
		return event
	}

	// Now scan forward from eventStartIdx to find where this event ends
	return detectEventFromIndex(lines, eventStartIdx)
}

// findMostRecentEventStart walks backward through the dataset to find the start of the most recent event
// Returns the index of the first bad ping in the event cluster, or -1 if no event found
func findMostRecentEventStart(lines []PingLine) int {
	if len(lines) < 2 {
		return -1
	}

	// Walk backward to find 2+ bad pings within 4 pings
	for i := len(lines) - 1; i >= 3; i-- {
		// Check the 4 pings ending at i
		badCount := 0
		for j := i - 3; j <= i; j++ {
			if lines[j].Status == "timeout" || lines[j].Status == "delayed" {
				badCount++
			}
		}

		// If we found 2+ bad pings in this window, we found an event
		if badCount >= 2 {
			// Now walk backward from here to find the START of this event cluster
			// Continue backward until we find 4 consecutive good pings or reach the beginning
			eventStartIdx := i - 3

			// Scan backward from eventStartIdx to find where the bad cluster starts
			for j := i; j >= 0; j-- {
				if lines[j].Status == "timeout" || lines[j].Status == "delayed" {
					eventStartIdx = j
				} else {
					// Check if we have 4 consecutive good pings before this point
					goodCount := 0
					for k := j; k >= 0 && goodCount < 4; k-- {
						if lines[k].Status == "ok" {
							goodCount++
						} else {
							break
						}
					}
					// If we found 4 consecutive good pings (or reached start), this is our boundary
					if goodCount >= 4 || j == 0 {
						break
					}
				}
			}

			return eventStartIdx
		}
	}

	return -1
}

// detectEventFromIndex scans forward from a given start index to detect the complete event
func detectEventFromIndex(lines []PingLine, startIdx int) EventStatus {
	event := EventStatus{}

	if startIdx < 0 || startIdx >= len(lines) {
		return event
	}

	// Set latest ping time
	latestTime, err := parsePingTime(lines[len(lines)-1].StartTime)
	if err == nil {
		event.LatestPingTime = latestTime
	}

	// Set the start time
	startTime, err := parsePingTime(lines[startIdx].StartTime)
	if err == nil {
		event.StartTime = startTime
	}

	// Scan forward to find the most recent bad ping and check if event is still active
	var mostRecentBadIdx int = -1
	var mostRecentBadTime time.Time

	for i := startIdx; i < len(lines); i++ {
		if lines[i].Status == "timeout" || lines[i].Status == "delayed" {
			mostRecentBadIdx = i
			t, err := parsePingTime(lines[i].StartTime)
			if err == nil {
				mostRecentBadTime = t
			}
		}
	}

	// If no bad pings found from start, this shouldn't happen, but handle it
	if mostRecentBadIdx == -1 {
		return event
	}

	event.LastBadPing = mostRecentBadTime

	// Check if 4 consecutive good pings follow the last bad ping
	if mostRecentBadIdx < len(lines)-1 {
		consecutiveGood := 0
		firstGoodPingIdx := -1

		for i := mostRecentBadIdx + 1; i < len(lines); i++ {
			if lines[i].Status == "ok" {
				if consecutiveGood == 0 {
					firstGoodPingIdx = i
				}
				consecutiveGood++
				if consecutiveGood >= 4 {
					// Event has ended
					event.IsActive = false
					if firstGoodPingIdx != -1 {
						endTime, err := parsePingTime(lines[firstGoodPingIdx].StartTime)
						if err == nil {
							event.EndTime = endTime
						}
					}

					// Calculate duration: from start to first of 4 ending good pings
					if !event.StartTime.IsZero() && !event.EndTime.IsZero() {
						event.Duration = event.EndTime.Sub(event.StartTime)
					}
					return event
				}
			} else {
				// Broke the streak
				break
			}
		}
	}

	// If we get here, the event is still active (no 4 consecutive good pings after last bad)
	event.IsActive = true
	if !event.StartTime.IsZero() && !event.LatestPingTime.IsZero() {
		event.Duration = event.LatestPingTime.Sub(event.StartTime)
	}

	return event
}

// parsePingTime parses both RFC3339 and custom timestamp formats
func parsePingTime(timeStr string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err == nil {
		return t, nil
	}

	// Try alternate format
	t, err = time.Parse("2006-01-02 15:04:05.000", timeStr)
	if err == nil {
		return t, nil
	}

	return time.Time{}, err
}
