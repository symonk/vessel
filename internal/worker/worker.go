package worker

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/symonk/vessel/internal/config"
	"github.com/symonk/vessel/internal/stats"
	"github.com/symonk/vessel/internal/trace"
)

// Worker is a struct that can accept requests to dispatch
// on it's input channel and forward results on for collection
// via it's outbound channel.
//
// Worker keeps an instance of TraceData for setting granularity
// on the request lifecycle, this avoids even utilising a sync.Pool
// as a worker can only be processing a single request at a time
// and all fields on the TraceData are set on a per request basis.
type Worker struct {
	client     *http.Client
	requestsCh <-chan *http.Request
	resultsCh  chan<- *stats.Stats
	trace      *httptrace.ClientTrace
	wg         *sync.WaitGroup
	root       context.Context // Avoid many heap allocs, use a shared root.
	cfg        *config.Config
}

// New instantiates a new worker and returns a ptr to
// the instance of it.
func New(client *http.Client, in <-chan *http.Request, out chan<- *stats.Stats, wg *sync.WaitGroup, root context.Context, cfg *config.Config) *Worker {
	return &Worker{
		client:     client,
		requestsCh: in,
		resultsCh:  out,
		trace:      new(httptrace.ClientTrace),
		wg:         wg,
		root:       root,
		cfg:        cfg,
	}
}

// Accept begins accepting requests until the internal request
// input channel is closed then it will exit gracefully.
func (w *Worker) Accept() {
	if w == nil {
		return
	}
	defer w.wg.Done()
	for {
		select {
		case request, ok := <-w.requestsCh:
			if !ok {
				return
			}
			trace := w.prepareTracer()
			response, began, err := w.send(request)
			w.report(trace, response, began, err)
		case <-w.root.Done():
			// signal interrupt
		}
	}
}

// send dispatches the request to the client.  This allows granular control
// of the context cancellation without having to handle stacking deferrals
// of cancel funcs in a loop elsewhere leading to a potential memory leak.
func (w *Worker) send(request *http.Request) (*http.Response, time.Time, error) {
	ctx, cancel := w.context(w.cfg.Duration)
	defer cancel()
	request = request.Clone(ctx)
	// TODO: Does this play nice with timing out ctx?
	request = request.WithContext(httptrace.WithClientTrace(w.root, w.trace))
	when := time.Now()
	response, err := w.client.Do(request)
	return response, when, err
}

// report publishes appropriate data for a downstream system to consume
// in order to make sense of results.
func (w *Worker) report(trace *trace.Trace, response *http.Response, began time.Time, err error) {
	s := new(stats.Stats)

	// TODO: Implement actual bytes capturing of the sent request.  The actual action
	// of sending it via the client however will drain the stream, so a copy is required
	// initially, there will be a memory penalty to pay there.
	s.BytesSent = 0

	// capture pre-body read latency, it will be overwritten if a response
	// body read occurs later.
	s.Latency = time.Since(began)
	// The request error'd, there likely is no response body.
	if err != nil {
		s.Err = err
		return
	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		s.Err = err
	}
	s.Latency = time.Since(began)
	s.BytesReceived = int64(len(bytes))

	// Bolt on trace analytics from the lifecycle
	s.TimeOnDns = trace.DnsDone
	s.TimeOnTls = trace.TlsDone
	s.TimeOnConn = trace.GotConnection
	s.TimeOnConnect = trace.ConnectDone

	for {
		select {
		case w.resultsCh <- s:
		case <-w.root.Done():
			return
		}
	}
}

// context returns a sensible context that honours the users timeout specific flags
func (w *Worker) context(duration time.Duration) (context.Context, context.CancelFunc) {
	if duration == 0 {
		return w.root, func() {
			// user has requested no timeout.
		}
	}
	return context.WithTimeout(w.root, duration)
}

// prepareTracer sets up the shared internal trace instance with fresh
// values for a particular request.
func (w *Worker) prepareTracer() *trace.Trace {
	trace := new(trace.Trace)
	w.trace.DNSStart = func(info httptrace.DNSStartInfo) {
		trace.DnsStart = time.Now()
	}
	w.trace.DNSDone = func(info httptrace.DNSDoneInfo) {
		trace.DnsDone = time.Since(trace.DnsStart)
	}
	w.trace.ConnectStart = func(network string, addr string) {
		trace.ConnectStart = time.Now()
	}
	w.trace.ConnectDone = func(network string, addr string, err error) {
		trace.ConnectDone = time.Since(trace.ConnectStart)
	}
	w.trace.TLSHandshakeStart = func() {
		trace.TlsStart = time.Now()
	}
	w.trace.TLSHandshakeDone = func(state tls.ConnectionState, err error) {
		trace.TlsDone = time.Since(trace.TlsStart)
	}
	w.trace.GetConn = func(hostPort string) {
		trace.GettingConnection = time.Now()
	}
	w.trace.GotConn = func(conn httptrace.GotConnInfo) {
		trace.GotConnection = time.Since(trace.GettingConnection)
		trace.ReusedConnection = conn.Reused
	}
	return trace
}
