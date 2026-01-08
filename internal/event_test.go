package internal

import (
	"testing"
	"time"
)

// Helper to create a ping line
func makePing(startTime string, status string, latency int64) PingLine {
	return PingLine{
		StartTime: startTime,
		EndTime:   startTime, // Simplified for testing
		Status:    status,
		Latency:   latency,
	}
}

func TestDetectEvent_NoEvent(t *testing.T) {
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "ok", 10),
		makePing("2026-01-08 16:00:15.000", "ok", 12),
		makePing("2026-01-08 16:00:30.000", "ok", 11),
		makePing("2026-01-08 16:00:45.000", "ok", 13),
	}

	event := DetectEvent(lines)

	if event.IsActive {
		t.Error("Expected no active event, got active event")
	}
	if !event.StartTime.IsZero() {
		t.Error("Expected zero StartTime, got:", event.StartTime)
	}
	if !event.EndTime.IsZero() {
		t.Error("Expected zero EndTime, got:", event.EndTime)
	}
}

func TestDetectEvent_SingleBadPing(t *testing.T) {
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "ok", 10),
		makePing("2026-01-08 16:00:15.000", "delayed", 150),
		makePing("2026-01-08 16:00:30.000", "ok", 11),
		makePing("2026-01-08 16:00:45.000", "ok", 13),
	}

	event := DetectEvent(lines)

	if event.IsActive {
		t.Error("Expected no active event with single bad ping, got active event")
	}
	if !event.StartTime.IsZero() {
		t.Error("Expected zero StartTime with single bad ping, got:", event.StartTime)
	}
}

func TestDetectEvent_TwoBadPingsInLast4_Active(t *testing.T) {
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "ok", 10),
		makePing("2026-01-08 16:00:15.000", "delayed", 150),
		makePing("2026-01-08 16:00:30.000", "ok", 11),
		makePing("2026-01-08 16:00:45.000", "delayed", 160),
	}

	event := DetectEvent(lines)

	if !event.IsActive {
		t.Error("Expected active event with 2 bad pings in last 4, got inactive")
	}

	// StartTime should be the older of the two bad pings
	startTime, _ := time.Parse("2006-01-02 15:04:05.000", "2026-01-08 16:00:15.000")
	if event.StartTime != startTime {
		t.Errorf("Expected StartTime %v, got %v", startTime, event.StartTime)
	}

	// Duration should be from StartTime to latest ping
	latestTime, _ := time.Parse("2006-01-02 15:04:05.000", "2026-01-08 16:00:45.000")
	expectedDuration := latestTime.Sub(startTime)
	if event.Duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, event.Duration)
	}
}

func TestDetectEvent_EventEnded_4GoodPings(t *testing.T) {
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "ok", 10),
		makePing("2026-01-08 16:00:15.000", "delayed", 150),
		makePing("2026-01-08 16:00:30.000", "timeout", 0),
		makePing("2026-01-08 16:00:45.000", "ok", 11), // Good ping 1 (END TIME)
		makePing("2026-01-08 16:01:00.000", "ok", 12), // Good ping 2
		makePing("2026-01-08 16:01:15.000", "ok", 13), // Good ping 3
		makePing("2026-01-08 16:01:30.000", "ok", 14), // Good ping 4 - triggers end
	}

	event := DetectEvent(lines)

	if event.IsActive {
		t.Error("Expected ended event (not active), got active")
	}

	// EndTime should be the first of the 4 good pings (16:00:45)
	endTime, _ := time.Parse("2006-01-02 15:04:05.000", "2026-01-08 16:00:45.000")
	if event.EndTime != endTime {
		t.Errorf("Expected EndTime %v, got %v", endTime, event.EndTime)
	}

	// StartTime should be the oldest bad ping in this cluster (delayed at 16:00:15)
	startTime, _ := time.Parse("2006-01-02 15:04:05.000", "2026-01-08 16:00:15.000")
	if event.StartTime != startTime {
		t.Errorf("Expected StartTime %v, got %v", startTime, event.StartTime)
	}

	// Duration should be from first bad ping to first of the 4 ending good pings
	expectedDuration := endTime.Sub(startTime)
	if event.Duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, event.Duration)
	}
}

func TestDetectEvent_EventEnded_3GoodPings_StillActive(t *testing.T) {
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "ok", 10),
		makePing("2026-01-08 16:00:15.000", "delayed", 150),
		makePing("2026-01-08 16:00:30.000", "timeout", 0),
		makePing("2026-01-08 16:00:45.000", "ok", 11),
		makePing("2026-01-08 16:01:00.000", "ok", 12),
		makePing("2026-01-08 16:01:15.000", "ok", 13),
		// Only 3 good pings after last bad - not enough to end event
	}

	event := DetectEvent(lines)

	// With new backward-walking logic, we find the most recent event cluster (the 2 bad pings)
	// Since there are only 3 good pings after them (not 4+), the event should be marked as active
	if !event.IsActive {
		t.Error("Expected active event (only 3 good pings after bad), got inactive")
	}

	// StartTime should be the first bad ping in the cluster
	startTime, _ := time.Parse("2006-01-02 15:04:05.000", "2026-01-08 16:00:15.000")
	if event.StartTime != startTime {
		t.Errorf("Expected StartTime %v, got %v", startTime, event.StartTime)
	}

	// Duration should be from start to latest ping
	latestTime, _ := time.Parse("2006-01-02 15:04:05.000", "2026-01-08 16:01:15.000")
	expectedDuration := latestTime.Sub(startTime)
	if event.Duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, event.Duration)
	}
}

func TestDetectEvent_MultipleEventClusters(t *testing.T) {
	lines := []PingLine{
		// First event cluster
		makePing("2026-01-08 16:00:00.000", "delayed", 150),
		makePing("2026-01-08 16:00:15.000", "timeout", 0),
		makePing("2026-01-08 16:00:30.000", "ok", 11),
		makePing("2026-01-08 16:00:45.000", "ok", 12),
		makePing("2026-01-08 16:01:00.000", "ok", 13),
		makePing("2026-01-08 16:01:15.000", "ok", 14), // First event ends here
		// Good period
		makePing("2026-01-08 16:01:30.000", "ok", 10),
		makePing("2026-01-08 16:01:45.000", "ok", 11),
		// Second event cluster
		makePing("2026-01-08 16:02:00.000", "delayed", 160),
		makePing("2026-01-08 16:02:15.000", "delayed", 170),
		makePing("2026-01-08 16:02:30.000", "ok", 12),
	}

	event := DetectEvent(lines)

	// Should have no active event (last 4 pings: ok, ok, delayed, delayed = 2 bad = active)
	// Actually, last 4 are: delayed(16:02:15), delayed(16:02:00), ok(16:01:45), ok(16:01:30)
	// Wait, I need to think about this more carefully. The last 4 pings from the end:
	// delayed(16:02:30), delayed(16:02:15), delayed(16:02:00), ok(16:01:45)
	// That's 3 bad pings in last 4, which > 2, so IS active

	if !event.IsActive {
		t.Error("Expected active event (2+ bad in last 4), got inactive")
	}

	// StartTime should be from the current event cluster
	startTime, _ := time.Parse("2006-01-02 15:04:05.000", "2026-01-08 16:02:00.000")
	if event.StartTime != startTime {
		t.Errorf("Expected StartTime %v, got %v", startTime, event.StartTime)
	}
}

func TestDetectEvent_EmptyList(t *testing.T) {
	lines := []PingLine{}

	event := DetectEvent(lines)

	if event.IsActive {
		t.Error("Expected no active event for empty list, got active")
	}
	if !event.StartTime.IsZero() {
		t.Error("Expected zero StartTime for empty list, got:", event.StartTime)
	}
}

func TestDetectEvent_SinglePing_Good(t *testing.T) {
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "ok", 10),
	}

	event := DetectEvent(lines)

	if event.IsActive {
		t.Error("Expected no active event with single good ping, got active")
	}
}

func TestDetectEvent_SinglePing_Bad(t *testing.T) {
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "delayed", 150),
	}

	event := DetectEvent(lines)

	if event.IsActive {
		t.Error("Expected no active event with single bad ping, got active")
	}
}

func TestDetectEvent_ExactlyTwoBadPingsInLast4(t *testing.T) {
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "ok", 10),
		makePing("2026-01-08 16:00:15.000", "delayed", 150),
		makePing("2026-01-08 16:00:30.000", "delayed", 160),
		makePing("2026-01-08 16:00:45.000", "ok", 11),
	}

	event := DetectEvent(lines)

	if !event.IsActive {
		t.Error("Expected active event (exactly 2 bad in last 4), got inactive")
	}
}

func TestDetectEvent_ThreeBadPingsInLast4(t *testing.T) {
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "ok", 10),
		makePing("2026-01-08 16:00:15.000", "delayed", 150),
		makePing("2026-01-08 16:00:30.000", "delayed", 160),
		makePing("2026-01-08 16:00:45.000", "timeout", 0),
	}

	event := DetectEvent(lines)

	if !event.IsActive {
		t.Error("Expected active event (3 bad in last 4), got inactive")
	}

	// StartTime should be oldest of the bad pings
	startTime, _ := time.Parse("2006-01-02 15:04:05.000", "2026-01-08 16:00:15.000")
	if event.StartTime != startTime {
		t.Errorf("Expected StartTime %v, got %v", startTime, event.StartTime)
	}
}

func TestDetectEvent_EventEndedButNoStaleData(t *testing.T) {
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "delayed", 150),
		makePing("2026-01-08 16:00:15.000", "timeout", 0),
		makePing("2026-01-08 16:00:30.000", "ok", 11), // First of 4 good pings
		makePing("2026-01-08 16:00:45.000", "ok", 12),
		makePing("2026-01-08 16:01:00.000", "ok", 13),
		makePing("2026-01-08 16:01:15.000", "ok", 14),
	}

	event := DetectEvent(lines)

	if event.IsActive {
		t.Error("Expected ended event, got active")
	}

	// EndTime should be the first of the 4 good pings (16:00:30)
	endTime, _ := time.Parse("2006-01-02 15:04:05.000", "2026-01-08 16:00:30.000")
	if event.EndTime != endTime {
		t.Errorf("Expected EndTime %v, got %v", endTime, event.EndTime)
	}

	// StartTime should be the first bad ping (16:00:00)
	startTime, _ := time.Parse("2006-01-02 15:04:05.000", "2026-01-08 16:00:00.000")
	expectedDuration := endTime.Sub(startTime)
	if event.Duration != expectedDuration {
		t.Errorf("Expected duration %v, got %v", expectedDuration, event.Duration)
	}
}

func TestDetectEvent_GoodPingsBrokenByBadPing(t *testing.T) {
	// 4 good pings, then a bad ping should break the end detection and keep event ended
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "delayed", 150),
		makePing("2026-01-08 16:00:15.000", "ok", 11),
		makePing("2026-01-08 16:00:30.000", "ok", 12),
		makePing("2026-01-08 16:00:45.000", "ok", 13),
		makePing("2026-01-08 16:01:00.000", "ok", 14),
		makePing("2026-01-08 16:01:15.000", "delayed", 160), // Breaks the good streak
	}

	event := DetectEvent(lines)

	// Last 4 pings: delayed(16:01:15), ok(16:01:00), ok(16:00:45), ok(16:00:30)
	// That's 1 bad ping, not enough for active
	if event.IsActive {
		t.Error("Expected no active event, got active")
	}
}

// Tests for fresh/small files
func TestDetectEvent_TwoPings_BothBad(t *testing.T) {
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "delayed", 150),
		makePing("2026-01-08 16:00:15.000", "timeout", 0),
	}

	event := DetectEvent(lines)

	// With backward-walking, we find bad pings even in small datasets
	// But since we need 4 pings to create a window, nothing is found in the backward loop
	// So we should have no event detected
	if event.IsActive {
		t.Error("Expected no active event with only 2 pings (can't form 4-ping window), got active")
	}
	if !event.StartTime.IsZero() {
		t.Error("Expected zero StartTime, got:", event.StartTime)
	}
}

func TestDetectEvent_ThreePings_TwoBad(t *testing.T) {
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "delayed", 150),
		makePing("2026-01-08 16:00:15.000", "timeout", 0),
		makePing("2026-01-08 16:00:30.000", "ok", 10),
	}

	event := DetectEvent(lines)

	// With backward-walking, we need 4 pings to form a window (i >= 3)
	// So nothing is found in the backward loop
	if event.IsActive {
		t.Error("Expected no active event with only 3 pings (can't form 4-ping window), got active")
	}
	if !event.StartTime.IsZero() {
		t.Error("Expected zero StartTime, got:", event.StartTime)
	}
}

func TestDetectEvent_FourPings_TwoBadAtEnd(t *testing.T) {
	// Fresh file with 4 pings, last 2 are bad
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "ok", 10),
		makePing("2026-01-08 16:00:15.000", "ok", 11),
		makePing("2026-01-08 16:00:30.000", "delayed", 150),
		makePing("2026-01-08 16:00:45.000", "timeout", 0),
	}

	event := DetectEvent(lines)

	if !event.IsActive {
		t.Error("Expected active event (2 bad in last 4), got inactive")
	}

	startTime, _ := time.Parse("2006-01-02 15:04:05.000", "2026-01-08 16:00:30.000")
	if event.StartTime != startTime {
		t.Errorf("Expected StartTime %v, got %v", startTime, event.StartTime)
	}
}

func TestDetectEvent_FourPings_AllGood(t *testing.T) {
	// Fresh file with all good pings
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "ok", 10),
		makePing("2026-01-08 16:00:15.000", "ok", 11),
		makePing("2026-01-08 16:00:30.000", "ok", 12),
		makePing("2026-01-08 16:00:45.000", "ok", 13),
	}

	event := DetectEvent(lines)

	if event.IsActive {
		t.Error("Expected no active event, got active")
	}
	if !event.StartTime.IsZero() {
		t.Error("Expected zero StartTime, got:", event.StartTime)
	}
}

func TestDetectEvent_FourPings_AllBad(t *testing.T) {
	// Fresh file with all bad pings (e.g., network down from start)
	lines := []PingLine{
		makePing("2026-01-08 16:00:00.000", "timeout", 0),
		makePing("2026-01-08 16:00:15.000", "timeout", 0),
		makePing("2026-01-08 16:00:30.000", "delayed", 150),
		makePing("2026-01-08 16:00:45.000", "timeout", 0),
	}

	event := DetectEvent(lines)

	if !event.IsActive {
		t.Error("Expected active event (multiple bad pings), got inactive")
	}

	startTime, _ := time.Parse("2006-01-02 15:04:05.000", "2026-01-08 16:00:00.000")
	if event.StartTime != startTime {
		t.Errorf("Expected StartTime %v, got %v", startTime, event.StartTime)
	}
}
