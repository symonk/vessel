package requester

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/symonk/vessel/internal/collector"
	"github.com/symonk/vessel/internal/config"
	"github.com/symonk/vessel/internal/trace"
)

type RequestResult struct {
	Response *http.Response
	Err      error
}

// Requester sends HTTP requests to a server (typically at scale) and
// can be signalled to wait until all requests have finalized through
// Wait()
type Requester interface {
	Wait()
}

// Requester takes a request and fans out many instances
// of that request until either the maximum count is reached
// or the duration has been surpassed.
type HTTPRequester struct {
	ctx       context.Context // Parent cancelled on signal
	collector collector.ResultCollector
	cfg       *config.Config
	client    *http.Client
	template  *http.Request
	workerCh  chan *http.Request
	wg        sync.WaitGroup
}

// New instantiates a new instance of Requester and returns
// the ptr to it.
func New(ctx context.Context, cfg *config.Config, collector collector.ResultCollector, template *http.Request) *HTTPRequester {
	maxWorkers := max(1, cfg.Concurrency)
	r := &HTTPRequester{
		ctx:       ctx,
		collector: collector,
		cfg:       cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
			Transport: NewRateLimitingTransport(
				cfg.MaxRPS,
				&RateLimitingTransport{
					// TODO: Overhaul this.
					Next: &http.Transport{
						Proxy:                 http.ProxyFromEnvironment,
						ForceAttemptHTTP2:     true,
						MaxConnsPerHost:       cfg.MaxConnections,
						IdleConnTimeout:       90 * time.Second,
						TLSHandshakeTimeout:   10 * time.Second,
						ExpectContinueTimeout: 1 * time.Second,
						TLSClientConfig: &tls.Config{
							// Skip server verification checks.  Enables testing against
							// self signed/expired certs, wrong domain or untrusted.
							InsecureSkipVerify: cfg.Insecure,
						},
					},
				}),
		},
		template: template,
		workerCh: make(chan *http.Request, maxWorkers),
	}
	r.wg.Add(maxWorkers)
	go r.spawn(maxWorkers)
	return r
}

// Wait waits until all requests are finished and all workers
// have cleanly shutdown.
func (h *HTTPRequester) Wait() {
	h.wg.Wait()
}

// spawn fans out workers in the pool upto the configured
// concurrency.
func (h *HTTPRequester) spawn(count int) {
	// Asynchronously start worker routines
	go func() {
		for range count {
			go worker(h.collector, &h.wg, h.client, h.workerCh)
		}
	}()

	// Asynchronously load requests into the queue.
	// Depending on -d or -a (duration || amount) the strategy
	// for loading requests onto the queues differs.
	var seen int64
	var tick <-chan time.Time
	if dur := h.cfg.Duration; dur > 0 {
		ticker := time.NewTicker(dur)
		defer ticker.Stop()
		tick = ticker.C
	}

	defer func() {
		close(h.workerCh)
	}()
	for {
		select {
		case <-tick:
			// if a duration was set, we have reached it.
			// gracefully exit.
			// nil channel otherwise (never selects/blocks)
			return
		case <-h.ctx.Done():
			// A signal was received, cause a graceful exit
			return
		default:
			// keep track of seen requests and keep providing requests
			// to workers as fast as possible.
			if tick == nil && seen == h.cfg.Amount {
				return
			}
			ctx, cancel := getCtx(h.cfg.Duration)
			defer cancel()
			r := h.template.Clone(ctx)

			// establish some trace vars
			traceData := new(trace.Trace)
			tracer := &httptrace.ClientTrace{
				DNSStart: func(info httptrace.DNSStartInfo) {
					traceData.DnsStart = time.Now()
				},
				DNSDone: func(info httptrace.DNSDoneInfo) {
					traceData.DnsDone = time.Since(traceData.DnsStart)
				},
				ConnectStart: func(network string, addr string) {
					traceData.ConnectStart = time.Now()
				},
				ConnectDone: func(network string, addr string, err error) {
					traceData.ConnectDone = time.Since(traceData.ConnectStart)
				},
				TLSHandshakeStart: func() {
					traceData.TlsStart = time.Now()
				},
				TLSHandshakeDone: func(state tls.ConnectionState, err error) {
					traceData.TlsDone = time.Since(traceData.TlsStart)
				},
				GotConn: func(conn httptrace.GotConnInfo) {
					if conn.Reused {
						traceData.ReusedConnection = true
					}
				},
			}

			// initialise the request with the tracing capabilities wrapped.
			// the client houses a total timeout provided on the command line
			// so a background ctx is sufficient here.  Store the tracing
			// information within it for later use.
			c := context.WithValue(context.Background(), trace.TraceDataKey, traceData)
			r = r.WithContext(httptrace.WithClientTrace(c, tracer))

			seen++
			h.workerCh <- r
		}
	}
}

func worker(collector collector.ResultCollector, wg *sync.WaitGroup, client *http.Client, work <-chan *http.Request) {
	defer wg.Done()
	for req := range work {
		// belt and braces, shouldn't be happening tho.
		if req == nil {
			continue
		}
		start := time.Now()

		response, err := client.Do(req)
		// TODO: Need to capture the request bytes sent over the wire to help paint a better
		// picture during summarisation of the throughput hueristics.
		// Always 0 for now.
		inspect(collector, response, time.Since(start), 0, err)
	}
}

// inspect encapsulates the logic of inspecting a single response.  This allows tidy
// closing of the response body etc on a per request request<->response basis rather than
// repeating the logic within the worker goroutine (cannot defer there reliably) and would
// create a sizable resource leak of responses at scale.
func inspect(collector collector.ResultCollector, response *http.Response, latency time.Duration, bytesSent int64, err error) {
	if err != nil {
		collector.Record(nil, latency, 0, bytesSent, err)
		return
	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		collector.Record(nil, latency, int64(len(bytes)), bytesSent, err)
		return
	}
	collector.Record(response, latency, int64(len(bytes)), bytesSent, nil)
}

func getCtx(duration time.Duration) (context.Context, context.CancelFunc) {
	if duration == 0 {
		return context.Background(), func() {}
	}
	return context.WithTimeout(context.Background(), duration)
}
