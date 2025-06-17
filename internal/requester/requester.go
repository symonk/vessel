package requester

import (
	"context"
	"net/http"

	"github.com/symonk/turbo"
	"github.com/symonk/vessel/internal/collector"
	"github.com/symonk/vessel/internal/config"
)

// Requester sends HTTP requests to a server (typically at scale) and
// can be signalled to wait until all requests have finalized through
// Wait()
type Requester interface {
	Do(request *http.Request) (*http.Response, error)
	Go(ctx context.Context)
	Wait()
}

// Requester takes a request and fans out many instances
// of that request until either the maximum count is reached
// or the duration has been surpassed.
type RequestSender struct {
	cfg      config.Config
	client   http.Client
	template *http.Request
	pool     turbo.Pooler
}

// New instantiates a new instance of Requester and returns
// the ptr to it.
func New(cfg config.Config, collector collector.Collector, template *http.Request) *RequestSender {
	return &RequestSender{
		cfg: cfg,
		client: http.Client{
			Timeout: cfg.Timeout,
			Transport: &RateLimitingTransport{
				Next: &CollectableTransport{
					Collector: collector,
					Next:      http.DefaultTransport,
				},
			},
		},
		template: template,
		pool:     turbo.NewPool(cfg.Concurrency),
	}
}

func (r *RequestSender) Go(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Either the application received an interrupt signal or the
			// user provided -d (duration flag has passed)
			return
		}

	}
}

// TODO: Fix error handling into turbo's stop()
func (r *RequestSender) Wait() {
	_ = r.pool.Stop(true)
}
