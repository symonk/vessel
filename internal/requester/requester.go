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
func (c *CollectableTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := c.Next.RoundTrip(request)
	return response, err
}

type Requester interface {
	Go()
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
			Transport: &CollectableTransport{
				Collector: collector,
				Next:      http.DefaultTransport,
			},
		},
		template: template,
	}
}

func (u *RequestSender) Go()   {}
func (u *RequestSender) Wait() {}
