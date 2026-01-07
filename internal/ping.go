package internal

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func spawnPing(target string, port int, useICMP bool, logFile string, verbose bool, wg *sync.WaitGroup) {
	defer wg.Done()

	startTime := time.Now()
	var latency time.Duration
	var status string
	var err error

	if useICMP {
		latency, err = IcmpPing(target)
	} else {
		latency, err = TcpPing(target, port, 15*time.Second)
	}

	if err != nil {
		status = "timeout"
	} else {
		status = classifyStatus(latency)
	}

	writeToCSV(logFile, startTime, time.Now(), latency, status)

	if verbose {
		fmt.Println(status)
	}
}

func TcpPing(target string, port int, timeout time.Duration) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", target, port), timeout)
	if err != nil {
		return 0, err
	}
	conn.Close()
	return time.Since(start), nil
}

func IcmpPing(target string) (time.Duration, error) {
	return 42 * time.Millisecond, nil // placeholder
}

func classifyStatus(latency time.Duration) string {
	if latency < 100*time.Millisecond {
		return "ok"
	}
	return "delayed"
}
