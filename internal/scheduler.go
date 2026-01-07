package internal

import (
	"fmt"
	"time"
)

func StartScheduler(cfg Config) {
	for {
		periodStart, periodEnd := calculatePeriod(cfg.DebugMode)
		logFile := prepareLogFile(periodStart)

		fmt.Printf("[Scheduler] New period: %v to %v\n", periodStart, periodEnd)

		for time.Now().Before(periodEnd) {
			nextPingTime := alignToSchedule(periodStart, cfg.Interval)
			spawnTime := maxTime(time.Now(), nextPingTime)

			sleepUntil(spawnTime)
			go spawnPing(cfg.Target, cfg.Port, cfg.UseICMP, logFile, cfg.Verbose)
		}

		waitForAllPings()
		generateSummary(logFile)
	}
}

func calculatePeriod(debug bool) (time.Time, time.Time) {
	now := time.Now()
	if debug {
		if now.Second() <= 30 {
			start := now.Truncate(time.Minute)
			return start, start.Add(30 * time.Second)
		}
		start := now.Truncate(time.Minute).Add(30 * time.Second)
		return start, start.Add(30 * time.Second)
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
