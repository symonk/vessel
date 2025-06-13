package collector

import (
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/symonk/vessel/internal/config"
)

// StatusBucket houses the breakdown of how many
// of each status code was received for all requests.
type StatusBucket map[int]int

// String implements fmt.Stringer and returns a formatted
// breakdown of responses.
func (s StatusBucket) String() string {
	var b strings.Builder
	b.WriteString("Response Breakdown:")
	for k, v := range s {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("\t[%d]: %d", k, v))
	}
	return b.String()

}

// Count returns the total count of responses in the map.
func (s StatusBucket) Count() int {
	var c int
	for k := range s {
		c += k
	}
	return c
}

type Collector interface {
	RecordSuccess(code int)
	RecordFailure(code int, err error)
	Summarise() string
}

// EventCollecter collects metrics of interest throughout
// the execution and can summarise those to an output
// stream.
type EventCollector struct {
	seen    atomic.Int64
	started time.Time
	cfg     *config.Config
	bucket  StatusBucket
}

func New(cfg *config.Config) *EventCollector {
	return &EventCollector{
		cfg:     cfg,
		started: time.Now(),
		bucket: StatusBucket{
			200: 10,
			204: 30,
			401: 100,
			403: 5,
			500: 19,
		},
	}
}

func (e *EventCollector) RecordSuccess(code int)            {}
func (e *EventCollector) RecordFailure(code int, err error) {}

// Summarise calculates the final summary prior to exiting.
// Complex logic will occur in here based on all the kinds
// of responses observed for the various requests sent.
func (e *EventCollector) Summarise() string {
	return fmt.Sprintf(`Running %s test against %s
%d connections.

Summary:
  Total Requests:	%d
  Duration: 		%s
  Latency:      	avg=8.3ms max=240ms p95=15ms
  Errors:       	2 timeouts, 3 connection resets
  Throughput:   	1.1MB/s
 
  %s`,
		e.cfg.Duration,
		e.cfg.Endpoint,
		e.cfg.Duration,
		e.bucket.Count(),
		time.Since(e.started),
		e.bucket,
	)
}
