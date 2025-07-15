package requester

import (
	"net/http"
	"time"

	"github.com/symonk/vessel/internal/collector"
)

// CollectingTransport enables publishing 'metrics' to
// the collector as part of request->response flow.
type CollectingTransport struct {
	Collector collector.ResultCollector
	Next      http.RoundTripper
}

// RoundTrip collects and publishes metrics to the collector for each individual
// request/response.
func (c *CollectingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	start := time.Now()
	response, err := c.Next.RoundTrip(request)
	latency := time.Since(start)
	// TODO: Figure out the mechanism for passing data through to the collector
	_ = latency
	c.Collector.Record(response, start, err)
	return response, err
}

// RateLimitingTransport allows limiting max RPS
// at the transport layer
type RateLimitingTransport struct {
	Next       http.RoundTripper
	sema       chan struct{}
	throttling bool
}

// NewRateLimitingTransport throttles requests so that only a fixed
// number may be in flight at a given time.  This prevents infinite
// goroutine scaling and potentially thrashing.
func NewRateLimitingTransport(maximum int, Next http.RoundTripper) *RateLimitingTransport {
	maximum = max(maximum, 0)
	var sema chan struct{}
	if maximum > 0 {
		sema = make(chan struct{}, maximum)
	}
	return &RateLimitingTransport{
		Next:       Next,
		sema:       sema,
		throttling: sema != nil,
	}

}

// RoundTrip enforces the rate limiter and forwards the request on
// through the request chains.
func (r *RateLimitingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if r.throttling {
		r.sema <- struct{}{}
		defer func() {
			<-r.sema
		}()
	}
	return r.Next.RoundTrip(request)
}
