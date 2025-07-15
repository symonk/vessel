package collector

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/symonk/vessel/internal/config"
)

// Recorder is the interface for something which records metrics from
// a HTTP Request->Response interactions.
type Recorder interface {
	Record(response *http.Response, sent time.Time, err error)
}

// Summariser is the interface for something which can display summary
// information.
type Summariser interface {
	Summarise()
}

type ResultCollector interface {
	Summariser
	Recorder
}

// EventCollector collects execution data during the lifecycle of
// vessell in order to build a meaningful summary.
//
// EventCollector captures information on the following metrics:
//
// - Latency (p50, p75, p90, p99)
// - Throughput
// - Errors
// - Response status code spreads
// - Indepth metrics throughout the HTTP lifecycle, including:
//
// - Time until first response byte
// - Time spent performing DNS lookups
// - Time spent in the TLS Handshake
// - Time spent managing connections
type EventCollector struct {
	counter          *StatusCodeCounter
	cfg              *config.Config
	writer           io.Writer
	started          time.Time
	errors           []error
	mu               sync.Mutex
	latencyHistogram hdrhistogram.Histogram
	bytesTransferred atomic.Int64
}

func New(writer io.Writer, cfg *config.Config) *EventCollector {
	return &EventCollector{
		counter:          NewStatusCodeCounter(),
		cfg:              cfg,
		writer:           writer,
		started:          time.Now(),
		latencyHistogram: *hdrhistogram.New(1, 60000, 3),
	}
}

// Record captures information about the completed request.
// It keeps mutex locking to a minimum where possible and favours
// CPU atomic operations where possible.
func (e *EventCollector) Record(response *http.Response, sent time.Time, err error) {
	// It is possible response is nil in error cases.
	// Keep a reference to the error, we will categorise them later
	// based on the different types.
	if err != nil {
		e.mu.Lock()
		defer e.mu.Unlock()
		e.errors = append(e.errors, err)
		return
	}

	// We have a semi-successful response (in that sense that no error was returned)
	// Capture the histogram data for the latency of the response.
	e.latencyHistogram.RecordValue(time.Since(sent).Milliseconds())
	e.counter.Increment(response.StatusCode)

	// Read the full response body to update the bytes received
	// TODO: This is experimental, probably can't do it here safely.
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	defer response.Body.Close() // TODO: This is dangerous for transports potentially later.
	e.bytesTransferred.Add(int64(len(bytes)))

	//

	//

	//
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
	total := e.counter.Count()
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
  Throughput:			{{.Throughput}}

  {{.Results}}

`
	// TODO: Smarter use of different terms, if the test was < 1MB transffered for example
	// fallback to bytes/sec etc etc.
	bytes := e.bytesTransferred.Load()
	bytesPerSecond := (bytes / int64(seconds))
	mbConsumed := float64(bytesPerSecond) / 1_000_000
	s := &Summary{
		Host:      e.cfg.Endpoint,
		Duration:  e.cfg.Duration.String(),
		Count:     e.latencyHistogram.TotalCount(),
		PerSecond: perSec,
		// TODO: Less than millisecond precision support.
		Latency:    latency,
		Throughput: fmt.Sprintf("Mbs=%.2f", mbConsumed),
		ErrorCount: len(e.errors),
		RealTime:   done,
		Results:    e.counter,
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
