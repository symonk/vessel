package collector

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/symonk/vessel/internal/config"
	"github.com/symonk/vessel/internal/trace"
)

// Recorder is the interface for something which records metrics from
// a HTTP Request->Response interactions.
type Recorder interface {
	Record(response *http.Response, latency time.Duration, receivedBytes int64, sentBytes int64, err error)
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
// - Latency (p90, p95, p99)
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
	counter              *StatusCodeCounter
	cfg                  *config.Config
	writer               io.Writer
	collectionRegistered time.Time
	rawErrors            error
	errGrouper           *ErrorGrouper
	mu                   sync.Mutex
	latency              hdrhistogram.Histogram
	bytesReceived        atomic.Int64
	bytesSent            atomic.Int64
	waitingDns           time.Duration
	waitingTls           time.Duration
	waitingConnect       time.Duration
	newConnections       atomic.Int64
	waitingGetConn       time.Duration
}

func New(writer io.Writer, cfg *config.Config) *EventCollector {
	return &EventCollector{
		counter:              NewStatusCodeCounter(),
		cfg:                  cfg,
		writer:               writer,
		collectionRegistered: time.Now(),
		latency:              *hdrhistogram.New(1, 60000, 3),
		rawErrors:            nil,
		errGrouper:           NewErrGrouper(),
	}
}

// Record captures information about the completed request.
// It keeps mutex locking to a minimum where possible and favours
// CPU atomic operations where possible.
func (e *EventCollector) Record(response *http.Response, latency time.Duration, bytesReceived int64, bytesSent int64, err error) {
	// It is possible response is nil in error cases.
	// Keep a reference to the error, we will categorise them later
	// based on the different types.
	if err != nil {
		e.mu.Lock()
		defer e.mu.Unlock()
		e.rawErrors = errors.Join(e.rawErrors, err)
		e.errGrouper.Record(err)
		// TODO: Error grouping for smarter summarising.
		// TODO: Implement a way to 'classify' the errors into appropriate groups.
		return
	}

	// Pull out 'trace' data from the requests to paint a better picture in the summary
	// of how time was spent from a granular point of view.
	v := response.Request.Context().Value(trace.TraceDataKey)
	t, ok := v.(*trace.Trace)
	if ok {
		e.mu.Lock()
		e.waitingDns += t.DnsDone
		e.waitingTls += t.TlsDone
		e.waitingConnect += t.ConnectDone
		e.waitingGetConn += t.GotConnection
		e.mu.Unlock()
	}

	// We have a semi-successful response (in that sense that no error was returned)
	// Capture the histogram data for the latency of the response.
	e.latency.RecordValue(latency.Milliseconds())
	e.counter.Increment(response.StatusCode)

	// Track the byte size of the initial request aswell as content type of
	// the response from the server.  The collector is not responsible for
	// reading the response, this should be handled elsewhere to ensure safety
	// of reading responses and avoiding attempting multiple reads etc.
	e.bytesReceived.Add(bytesReceived)
	e.bytesSent.Add(bytesSent)

	// Keep track of keep-alives etc, useful for detecting if there is an issue
	// with your server, or our client.
	if !t.ReusedConnection {
		e.newConnections.Add(1)
	}
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
	done := time.Since(e.collectionRegistered)
	seconds := max(1, int(e.cfg.Duration.Seconds()))
	total := e.counter.Count()
	perSec := total / seconds
	latency := fmt.Sprintf("max=%dms, avg=%fms, p50=%dms, p90=%dms, p95=%dms, p99=%dms",
		e.latency.Max(),
		e.latency.Mean(),
		e.latency.ValueAtQuantile(50),
		e.latency.ValueAtQuantile(90),
		e.latency.ValueAtQuantile(95),
		e.latency.ValueAtQuantile(99),
	)

	// TODO: Be smarter here, capture terminal width and size appropriately.
	const tmpl = `
 _   _                    _ 
| | | |			 | |
| | | | ___  ___ ___  ___| |
| | | |/ ⚡\/ __/ __|/ ⚡\ |
\ \_/ /  __/\__ \__ \  __/ |
 \___/ \___||___/___/\___|_| https://github.com/symonk/vessel
                            
Running test @ {{.Host}} [vessel-{{.Version}}]
Workers: {{.Workers}}
Cores: {{.MaxProcs}}

complete [⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡⚡] {{.RealTime}}

Requests:	{{.Count}} ({{.PerSecond}}/second)
bytes:		Received({{.BytesReceived}}) | Sent({{.BytesSent}}) | Total({{.BytesTotal}})
Latency:	{{.Latency}}
Errored:	{{.Errors}}
Conns:		{{.OpenedConnections}}
TimeSpent:	{{.Waiting}}

{{.Results}}
`
	// TODO: Smarter use of different terms, if the test was < 1MB transffered for example
	// fallback to bytesReceived/sec etc etc.
	bytesReceived := e.bytesReceived.Load()
	receivedBytesPerSecond := (bytesReceived / int64(seconds))
	receivedMegabytes := float64(receivedBytesPerSecond) / 1_000_000

	bytesSent := e.bytesSent.Load()
	sentBytesPerSecond := (bytesSent / int64(seconds))
	sentMegabytes := float64(sentBytesPerSecond) / 1_000_000

	bytesTotal := receivedMegabytes + sentMegabytes

	// calculate granular breakdowns
	waitDns := e.waitingDns.Seconds()
	waitTls := e.waitingTls.Seconds()
	waitConnect := e.waitingConnect.Seconds()
	waitGettingConn := e.waitingGetConn.Seconds()

	totalDuration := max(e.cfg.Duration.Seconds(), 1)

	dnsPercent := (waitDns / totalDuration) * 100
	tlsPercent := (waitTls / totalDuration) * 100
	connectPercent := (waitConnect / totalDuration) * 100
	gettingConnPercent := (waitGettingConn / totalDuration) * 100

	waiting := fmt.Sprintf("[%.2f%%] Resolving DNS (%.2fs), [%.2f%%] TLS Handshake (%.2fs), [%.2f%%] Connecting (%.2fs) [%.2f%%] Getting Connections (%.2fs)",
		dnsPercent, waitDns/float64(e.cfg.Concurrency),
		tlsPercent, waitTls/float64(e.cfg.Concurrency),
		connectPercent, waitConnect/float64(e.cfg.Concurrency),
		gettingConnPercent, waitGettingConn/float64(e.cfg.Concurrency),
	)

	s := &Summary{
		Host:      e.cfg.Endpoint,
		Duration:  e.cfg.Duration.String(),
		Count:     e.latency.TotalCount(),
		PerSecond: perSec,
		// TODO: Less than millisecond precision support.
		Latency:           latency,
		BytesReceived:     fmt.Sprintf("%.2fMB", receivedMegabytes),
		BytesSent:         fmt.Sprintf("%.2fMB", sentMegabytes),
		RawErrors:         e.rawErrors,
		Errors:            e.errGrouper.String(),
		RealTime:          done,
		Results:           e.counter,
		Workers:           e.cfg.Concurrency,
		Version:           e.cfg.Version,
		Waiting:           waiting,
		OpenedConnections: e.newConnections.Load(),
		MaxProcs:          runtime.GOMAXPROCS(0),
		BytesTotal:        fmt.Sprintf("%.2FMB", bytesTotal),
	}
	t, err := template.New("summary").Parse(tmpl)
	if err != nil {
		// TODO: Improve
		fmt.Println("unable to generate summary")
	}
	outErr := t.Execute(e.writer, s)
	if outErr != nil {
		// TODO: Improve
		fmt.Println()
		fmt.Println("unable to show summary", outErr)
	}
}
