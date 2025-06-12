package requester

import (
	"net/http"
	"time"

	"github.com/symonk/vessel/internal/collector"
)

// CollectableTransport enables publishing 'metrics' to
// the collector as part of request->response flow.
type CollectableTransport struct {
	Collector collector.Collector
	Next      http.RoundTripper
}

// RoundTrip collects and publishes metrics to the collector for each individual
// request/response.
func (p *CollectableTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return nil, nil
}

type Requester interface {
	Go()
	Wait()
}

// Requester takes a request and fans out many instances
// of that request until either the maximum count is reached
// or the duration has been surpassed.
type RequestSender struct {
	client http.Client
}

// New instantiates a new instance of Requester and returns
// the ptr to it.
func New(collector collector.Collector, timeout time.Duration) *RequestSender {
	return &RequestSender{
		client: http.Client{
			Timeout: timeout,
			Transport: &CollectableTransport{
				Collector: collector,
				Next:      http.DefaultTransport,
			},
		},
	}
}

func (u *RequestSender) Go()   {}
func (u *RequestSender) Wait() {}
