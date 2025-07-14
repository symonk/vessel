package collector

import (
	"fmt"
	"html/template"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/symonk/vessel/internal/config"
)

// StatusBucket houses the breakdown of how many
// of each status code was received for all requests.
type StatusBucket map[int]int

// String implements fmt.Stringer and returns a formatted
// breakdown of responses.
func (s StatusBucket) String() string {
	var b strings.Builder
	b.WriteString("Response Breakdown:\n")
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
	RecordSuccess(code int, latencyMilliseconds int64)
	RecordFailure(err error)
	Summarise()
}

// EventCollecter collects metrics of interest throughout
// the execution and can summarise those to an output
// stream.
type EventCollector struct {
	writer           io.Writer
	started          time.Time
	cfg              *config.Config
	bucket           StatusBucket
	mu               sync.Mutex
	errors           []error
	latencyHistogram hdrhistogram.Histogram
}

func New(writer io.Writer, cfg *config.Config) *EventCollector {
	return &EventCollector{
		writer:           writer,
		cfg:              cfg,
		started:          time.Now(),
		bucket:           make(StatusBucket),
		latencyHistogram: *hdrhistogram.New(1, 60000, 3),
	}
}

// RecordSuccess atomically records a success
func (e *EventCollector) RecordSuccess(code int, latency int64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.latencyHistogram.RecordValue(latency)
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

	seconds := max(1, int(e.cfg.Duration.Seconds()))
	total := e.bucket.Count()
	perSec := total / seconds
	latency := fmt.Sprintf("max=%dms, avg=%fms, p50=%dms, p75=%dms, p95=%dms, p99=%dms, p99.9=%dms",
		e.latencyHistogram.Max(),
		e.latencyHistogram.Mean(),
		e.latencyHistogram.ValueAtQuantile(50),
		e.latencyHistogram.ValueAtQuantile(75),
		e.latencyHistogram.ValueAtQuantile(90),
		e.latencyHistogram.ValueAtQuantile(99),
		e.latencyHistogram.ValueAtQuantile(99.9),
	)

	const tmpl = `

Test against {{.Host}} finished after {{.RealTime}}.

Summary:

  Total Requests:		{{.Count}} ({{.PerSecond}} per second)
  Duration:			{{.RealTime}}
  Latency:			{{.Latency}}
  Errors:			{{.ErrorCount}}
  Throughput:			1.1MB/s

  {{.Results}}

`
	s := &Summary{
		Host:       e.cfg.Endpoint,
		Duration:   e.cfg.Duration.String(),
		Count:      e.latencyHistogram.TotalCount(),
		PerSecond:  perSec,
		Latency:    latency,
		Throughput: "1.1MB/s",
		ErrorCount: len(e.errors),
		RealTime:   done,
		Results:    e.bucket,
	}
	t, err := template.New("summary").Parse(tmpl)
	if err != nil {
		// TODO: Improve
		fmt.Println("unable to generate summary")
	}
	outErr := t.Execute(e.writer, s)
	if outErr != nil {
		// TODO: Improve
		fmt.Println("unable to show summary", err)
	}
}

type Summary struct {
	Host       string
	Duration   string
	Count      int64
	PerSecond  int
	Latency    string
	Throughput string
	ErrorCount int
	RealTime   time.Duration
	Results    StatusBucket
}
