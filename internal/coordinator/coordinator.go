package coordinator

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"github.com/symonk/vessel/internal/collector"
	"github.com/symonk/vessel/internal/config"
	"github.com/symonk/vessel/internal/stats"
	"github.com/symonk/vessel/internal/worker"
)

// Coordinator sends HTTP requests to a server (typically at scale) and
// can be signalled to wait until all requests have finalized through
// Wait()
type Coordinator interface {
	Wait()
}

// RequestCoordinator takes a request and fans out many instances
// of that request until either the maximum count is reached
// or the duration has been surpassed.
type RequestCoordinator struct {
	ctx       context.Context // Parent cancelled on signal
	collector collector.ResultCollector
	out       chan<- *stats.Stats
	cfg       *config.Config
	client    *http.Client
	template  *http.Request
	workerCh  chan *http.Request
	wg        sync.WaitGroup
}

// New instantiates a new instance of RequestCoordinator and returns
// the ptr to it.
func New(ctx context.Context, out chan<- *stats.Stats, cfg *config.Config, collector collector.ResultCollector, template *http.Request) *RequestCoordinator {
	maxWorkers := max(1, cfg.Concurrency)
	r := &RequestCoordinator{
		ctx:       ctx,
		collector: collector,
		cfg:       cfg,
		out:       out,
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
func (r *RequestCoordinator) Wait() {
	r.wg.Wait()
}

// spawn fans out workers in the pool upto the configured
// concurrency.
func (r *RequestCoordinator) spawn(count int) {
	for range count {
		w := worker.New(r.client, r.workerCh, r.out, &r.wg, r.ctx, r.cfg)
		go w.Accept()
	}

	// Asynchronously load requests into the queue.
	// Depending on -d or -a (duration || amount) the strategy
	// for loading requests onto the queues differs.
	var seen int64
	var tick <-chan time.Time
	if dur := r.cfg.Duration; dur > 0 {
		ticker := time.NewTicker(dur)
		defer ticker.Stop()
		tick = ticker.C
	}

	defer func() {
		close(r.workerCh)
	}()
	for {
		select {
		case <-tick:
			// if a duration was set, we have reached it.
			// gracefully exit.
			// nil channel otherwise (never selects/blocks)
			return
		case <-r.ctx.Done():
			// A signal was received, cause a graceful exit
			return
		default:
			// keep track of seen requests and keep providing requests
			// to workers as fast as possible.
			if tick == nil && seen == r.cfg.Amount {
				return
			}
			seen++
			r.workerCh <- r.template
		}
	}
}
