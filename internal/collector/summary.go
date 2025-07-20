package collector

import "time"

type Summary struct {
	Host          string
	Version       string
	Duration      string
	Count         int64
	PerSecond     int
	Latency       string
	BytesReceived string
	BytesSent     string
	ErrorCount    int
	RealTime      time.Duration
	Results       *StatusCodeCounter
	Connections   int
}
