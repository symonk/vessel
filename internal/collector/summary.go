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
	// TODO: This will likely go away, debug aid for now while building the error grouper.
	// This is a single 'joined' error for now.
	RawErrors   error
	Errors      string
	RealTime    time.Duration
	Results     *StatusCodeCounter
	Connections int
	Waiting     string
}
