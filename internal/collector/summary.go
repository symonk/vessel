package collector

import "time"

type Summary struct {
	Host       string
	Duration   string
	Count      int64
	PerSecond  int
	Latency    string
	Throughput string
	ErrorCount int
	RealTime   time.Duration
	Results    *StatusCodeCounter
}
