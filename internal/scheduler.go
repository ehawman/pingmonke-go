package internal

import (
	"fmt"
	"sync"
	"time"
)

func StartScheduler(cfg Config) {
	// Set interval based on debug mode
	interval := cfg.Interval
	if cfg.DebugMode {
		interval = cfg.DebugInterval
	}

	for {
		periodStart, periodEnd := calculatePeriod(cfg)
		logFile := prepareLogFile(periodStart, cfg)

		fmt.Printf("[Scheduler] New period: %v to %v\n", periodStart, periodEnd)

		var wg sync.WaitGroup

		for time.Now().Before(periodEnd) {
			nextPingTime := alignToSchedule(periodStart, interval)
			spawnTime := maxTime(time.Now(), nextPingTime)

			sleepUntil(spawnTime)
			wg.Add(1)
			go spawnPing(cfg.Target, cfg.Port, cfg.UseICMP, logFile, cfg.Verbose, &wg)
		}

		// Wait for all pings in this period to complete
		wg.Wait()
		generateSummary(logFile, cfg)

		// Sleep until next period to avoid busy loop at rollover
		sleepUntil(periodEnd.Add(100 * time.Millisecond))
	}
}

func calculatePeriod(cfg Config) (time.Time, time.Time) {
	now := time.Now()
	if cfg.DebugMode {
		start := now.Truncate(time.Minute)
		return start, start.Add(60 * time.Second)
	}
	start := time.Now().Truncate(24 * time.Hour)
	return start, start.Add(24 * time.Hour)
}

func alignToSchedule(base time.Time, interval time.Duration) time.Time {
	elapsed := time.Since(base)
	n := int(elapsed.Seconds()) / int(interval.Seconds())
	return base.Add(time.Duration(n+1) * interval)
}

func sleepUntil(t time.Time) {
	time.Sleep(time.Until(t))
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
