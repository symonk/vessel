package collector

import (
	"fmt"
	"io"
	"strings"
	"sync"
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
		c += s[k]
	}
	return c
}

type ResultCollector interface {
	RecordSuccess(code int)
	RecordFailure(err error)
	Summarise()
}

// EventCollecter collects metrics of interest throughout
// the execution and can summarise those to an output
// stream.
type EventCollector struct {
	writer  io.Writer
	started time.Time
	cfg     *config.Config
	bucket  StatusBucket
	mu      sync.Mutex
	errors  []error
}

func New(writer io.Writer, cfg *config.Config) *EventCollector {
	return &EventCollector{
		writer:  writer,
		cfg:     cfg,
		started: time.Now(),
		bucket:  make(StatusBucket),
	}
}

// RecordSuccess atomically records a success
func (e *EventCollector) RecordSuccess(code int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	_, ok := e.bucket[code]
	if !ok {
		e.bucket[code] = 1
		return
	}
	e.bucket[code] += 1
}

func (e *EventCollector) RecordFailure(err error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.errors = append(e.errors, err)
}

// Summarise calculates the final summary prior to exiting.
// Complex logic will occur in here based on all the kinds
// of responses observed for the various requests sent.
// TODO: Use some sort of templating here?
// TODO: Group errors by 'type of error', show counts
// TODO: Don't just dump every single error
// TODO: Wire in throughput support
// TODO: Wire in latency support
// TODO: Wire in latency breakdowns from httptrace for:
// TODO: DNS resolution, TCP connection time, TLS handshake time, Time to first byte, total response time.
func (e *EventCollector) Summarise() {
	done := time.Since(e.started)
	reasons := make([]string, len(e.errors))
	for i, e := range e.errors {
		reasons[i] = e.Error()
	}
	_, _ = fmt.Fprintf(e.writer, `Test against %s finished after %s.

Summary:
  Total Requests:	%d (%d per second)
  Duration: 		%s
  Latency:      	avg=8.3ms max=240ms p95=15ms
  Errors:       	%d
  Throughput:   	1.1MB/s

  ------------------------------------------------------------
 
  %s
  `,
		e.cfg.Endpoint,
		e.cfg.Duration.String(),
		e.bucket.Count(),
		e.bucket.Count()/int(e.cfg.Duration.Seconds()),
		done,
		len(e.errors),
		e.bucket,
	)
}
