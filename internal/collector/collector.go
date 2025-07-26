package collector

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"runtime"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/symonk/vessel/internal/config"
	"github.com/symonk/vessel/internal/stats"
)

// Summariser is the interface for something which can display summary
// information.
type Summariser interface {
	Summarise()
}

type ResultCollector interface {
	Summariser
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
	latency              hdrhistogram.Histogram
	bytesReceived        int64
	bytesSent            int64
	waitingDns           time.Duration
	waitingTls           time.Duration
	waitingConnect       time.Duration
	newConnections       int64
	waitingGetConn       time.Duration
	ingress              chan *stats.Stats
}

func New(ingress chan *stats.Stats, writer io.Writer, cfg *config.Config) *EventCollector {
	e := &EventCollector{
		counter:              NewStatusCodeCounter(),
		cfg:                  cfg,
		writer:               writer,
		collectionRegistered: time.Now(),
		latency:              *hdrhistogram.New(1, 60000, 3),
		rawErrors:            nil,
		errGrouper:           NewErrGrouper(),
		ingress:              ingress,
	}
	go e.listen()
	return e
}

// listen waits for stats from the worker pool before incremental internal
// values in preparation for summary generation later.
//
// For now this is a single listener, but eventually the channel can be fanned
// out for reads and merged back into a single result chan for efficiency.
func (e *EventCollector) listen() {
	fmt.Println("listening for stats")
	for stat := range e.ingress {
		err := stat.Err
		if err != nil {
			e.rawErrors = errors.Join(e.rawErrors, err)
			e.errGrouper.Record(err)
		}
		e.waitingDns += stat.TimeOnDns
		e.waitingTls += stat.TimeOnTls
		e.waitingConnect += stat.TimeOnConnect
		e.waitingGetConn += stat.TimeOnConn

		// We have a semi-successful response (in that sense that no error was returned)
		// Capture the histogram data for the latency of the response.
		e.latency.RecordValue(stat.Latency.Milliseconds())
		e.counter.Increment(stat.StatusCode)

		// Track the byte size of the initial request aswell as content type of
		// the response from the server.  The collector is not responsible for
		// reading the response, this should be handled elsewhere to ensure safety
		// of reading responses and avoiding attempting multiple reads etc.
		e.bytesReceived += stat.BytesReceived
		e.bytesSent += stat.BytesSent

		// Keep track of keep-alives etc, useful for detecting if there is an issue
		// with your server, or our client.
		e.newConnections += stat.ReusedConn
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
Waiting:	{{.Waiting}}

{{.Results}}

{{.RawErrors}}
`
	// TODO: Smarter use of different terms, if the test was < 1MB transffered for example
	// fallback to bytesReceived/sec etc etc.
	bytesReceived := e.bytesReceived
	receivedBytesPerSecond := (bytesReceived / int64(seconds))
	receivedMegabytes := float64(receivedBytesPerSecond) / 1_000_000

	bytesSent := e.bytesSent
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
		OpenedConnections: e.newConnections,
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
