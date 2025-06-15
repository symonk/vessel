package requester

import (
	"context"
	"net/http"
	"time"

	"github.com/symonk/vessel/internal/collector"
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
	client   http.Client
	template *http.Request
}

// New instantiates a new instance of Requester and returns
// the ptr to it.
func New(collector collector.Collector, timeout time.Duration, template *http.Request) *RequestSender {
	return &RequestSender{
		client: http.Client{
			Timeout: timeout,
			Transport: &RateLimitingTransport{
				Next: &CollectableTransport{
					Collector: collector,
					Next:      http.DefaultTransport,
				},
			},
		},
		template: template,
	}
}

func (u *RequestSender) Go(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// Either the application received an interrupt signal or the
			// user provided -d (duration flag has passed)
			return
		}
	}
}
func (u *RequestSender) Wait() {}
