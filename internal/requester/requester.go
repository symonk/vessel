package requester

import (
	"net/http"
	"sync"

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
	cfg      config.Config
	client   *http.Client
	template *http.Request
	workerCh chan *http.Request
	results  chan RequestResult
	wg       sync.WaitGroup
}

// New instantiates a new instance of Requester and returns
// the ptr to it.
func New(cfg config.Config, collector collector.Collector, template *http.Request) *HTTPRequester {
	maxWorkers := max(1, cfg.Concurrency)
	r := &HTTPRequester{
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
		results:  make(chan RequestResult, maxWorkers),
	}
	r.wg.Add(maxWorkers)
	go r.spawn(maxWorkers)
	return r
}

// Wait waits until all requests are finished.
func (r *HTTPRequester) Wait() {
	r.wg.Wait()
}

// spawn fans out workers in the pool upto the configured
// concurrency.
func (r *HTTPRequester) spawn(count int) {
	for range count {
		go worker(&r.wg, r.client, r.workerCh, r.results)
	}

}

func worker(wg *sync.WaitGroup, client *http.Client, work <-chan *http.Request, results chan<- RequestResult) {
	defer wg.Done()
	for req := range work {
		response, err := client.Do(req)
		results <- RequestResult{
			Response: response,
			Err:      err,
		}
	}
}
