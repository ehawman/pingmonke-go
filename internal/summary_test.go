package internal

import (
	"testing"
	"time"
)

// TestEventDetectionDebugMode tests event detection with 15 second windows
func TestEventDetectionDebugMode(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		pings          []PingRecord
		expectedEvents int
	}{
		{
			name: "No events - all good pings",
			pings: []PingRecord{
				{StartTime: baseTime, Status: "ok", Latency: 50},
				{StartTime: baseTime.Add(5 * time.Second), Status: "ok", Latency: 50},
				{StartTime: baseTime.Add(10 * time.Second), Status: "ok", Latency: 50},
			},
			expectedEvents: 0,
		},
		{
			name: "Single bad ping - no event",
			pings: []PingRecord{
				{StartTime: baseTime, Status: "ok", Latency: 50},
				{StartTime: baseTime.Add(5 * time.Second), Status: "timeout", Latency: 0},
				{StartTime: baseTime.Add(10 * time.Second), Status: "ok", Latency: 50},
			},
			expectedEvents: 0,
		},
		{
			name: "Two bad pings within 15s - one event",
			pings: []PingRecord{
				{StartTime: baseTime, Status: "ok", Latency: 50},
				{StartTime: baseTime.Add(5 * time.Second), Status: "timeout", Latency: 0},
				{StartTime: baseTime.Add(10 * time.Second), Status: "timeout", Latency: 0},
				{StartTime: baseTime.Add(15 * time.Second), Status: "ok", Latency: 50},
				{StartTime: baseTime.Add(20 * time.Second), Status: "ok", Latency: 50},
			},
			expectedEvents: 1,
		},
		{
			name: "High latency counts as bad",
			pings: []PingRecord{
				{StartTime: baseTime, Status: "delayed", Latency: 100},
				{StartTime: baseTime.Add(5 * time.Second), Status: "delayed", Latency: 150},
				{StartTime: baseTime.Add(10 * time.Second), Status: "ok", Latency: 50},
			},
			expectedEvents: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := detectEvents(tt.pings, true) // debug mode = true
			if len(events) != tt.expectedEvents {
				t.Errorf("Expected %d events, got %d", tt.expectedEvents, len(events))
			}
		})
	}
}

// TestEventDetectionNormalMode tests event detection with 60 second windows
func TestEventDetectionNormalMode(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name           string
		pings          []PingRecord
		expectedEvents int
	}{
		{
			name: "No events - all good pings",
			pings: []PingRecord{
				{StartTime: baseTime, Status: "ok", Latency: 50},
				{StartTime: baseTime.Add(10 * time.Second), Status: "ok", Latency: 50},
				{StartTime: baseTime.Add(20 * time.Second), Status: "ok", Latency: 50},
			},
			expectedEvents: 0,
		},
		{
			name: "Two bad pings within 60s - one event",
			pings: []PingRecord{
				{StartTime: baseTime, Status: "ok", Latency: 50},
				{StartTime: baseTime.Add(30 * time.Second), Status: "timeout", Latency: 0},
				{StartTime: baseTime.Add(40 * time.Second), Status: "timeout", Latency: 0},
				{StartTime: baseTime.Add(50 * time.Second), Status: "ok", Latency: 50},
				{StartTime: baseTime.Add(60 * time.Second), Status: "ok", Latency: 50},
			},
			expectedEvents: 1,
		},
		{
			name: "Two bad pings more than 60s apart - no event",
			pings: []PingRecord{
				{StartTime: baseTime, Status: "timeout", Latency: 0},
				{StartTime: baseTime.Add(70 * time.Second), Status: "timeout", Latency: 0},
				{StartTime: baseTime.Add(80 * time.Second), Status: "ok", Latency: 50},
			},
			expectedEvents: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := detectEvents(tt.pings, false) // debug mode = false
			if len(events) != tt.expectedEvents {
				t.Errorf("Expected %d events, got %d", tt.expectedEvents, len(events))
			}
		})
	}
}

// TestIsBadPing tests the isBadPing helper function
func TestIsBadPing(t *testing.T) {
	tests := []struct {
		name  string
		ping  PingRecord
		isBad bool
	}{
		{
			name:  "Good ping - ok status",
			ping:  PingRecord{Status: "ok", Latency: 50},
			isBad: false,
		},
		{
			name:  "Good ping - delayed status below 100ms",
			ping:  PingRecord{Status: "delayed", Latency: 99},
			isBad: false,
		},
		{
			name:  "Bad ping - timeout status",
			ping:  PingRecord{Status: "timeout", Latency: 0},
			isBad: true,
		},
		{
			name:  "Bad ping - latency at 100ms",
			ping:  PingRecord{Status: "delayed", Latency: 100},
			isBad: true,
		},
		{
			name:  "Bad ping - latency above 100ms",
			ping:  PingRecord{Status: "delayed", Latency: 150},
			isBad: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBadPing(tt.ping)
			if result != tt.isBad {
				t.Errorf("Expected isBadPing to be %v, got %v", tt.isBad, result)
			}
		})
	}
}
