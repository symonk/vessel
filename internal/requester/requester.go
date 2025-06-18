package requester

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/symonk/vessel/internal/collector"
	"github.com/symonk/vessel/internal/config"
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
	ctx      context.Context // Parent cancelled on signal
	cfg      config.Config
	client   *http.Client
	template *http.Request
	workerCh chan *http.Request
	wg       sync.WaitGroup
}

// New instantiates a new instance of Requester and returns
// the ptr to it.
func New(ctx context.Context, cfg config.Config, collector collector.Collector, template *http.Request) *HTTPRequester {
	maxWorkers := max(1, cfg.Concurrency)
	r := &HTTPRequester{
		ctx: ctx,
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
			Transport: &RateLimitingTransport{
				Next: &CollectableTransport{
					Collector: collector,
					Next:      http.DefaultTransport,
				},
			},
		},
		template: template,
		workerCh: make(chan *http.Request, maxWorkers),
	}
	r.wg.Add(maxWorkers)
	go r.spawn(maxWorkers)
	return r
}

// Wait waits until all requests are finished.
func (h *HTTPRequester) Wait() {
	h.wg.Wait()
}

// spawn fans out workers in the pool upto the configured
// concurrency.
func (h *HTTPRequester) spawn(count int) {
	// Asynchronously start worker routines
	go func() {
		for range count {
			go worker(&h.wg, h.client, h.workerCh)
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

	go func() {
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
				if tick != nil && seen == h.cfg.Amount {
					return
				}
				ctx, cancel := context.WithTimeout(context.Background(), h.cfg.Timeout)
				defer cancel()
				r := h.template.Clone(ctx)
				seen++
				h.workerCh <- r
			}
		}
	}()
}

func worker(wg *sync.WaitGroup, client *http.Client, work <-chan *http.Request) {
	defer wg.Done()
	for req := range work {
		client.Do(req)
	}
}
