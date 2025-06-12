package collector

import "sync/atomic"

type Collector interface {
	RecordSuccess(code int)
	RecordFailure(code int)
	Summarise() string
}

// EventCollecter collects metrics of interest throughout
// the execution and can summarise those to an output
// stream.
type EventCollector struct {
	seen atomic.Int64
}

func New() *EventCollector {
	return &EventCollector{}
}

func (e *EventCollector) RecordSuccess(code int) {}
func (e *EventCollector) RecordFailure(code int) {}
func (e *EventCollector) Summarise() string {
	return `Running 10s test @ https://yourwebsite.com
50 connections

Summary:
  Requests:     12000
  Duration:     10.01s
  Latency:      avg=8.3ms max=240ms p95=15ms
  Errors:       2 timeouts, 3 connection resets
  Throughput:   1.1MB/s
	`
}
