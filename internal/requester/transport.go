package requester

import (
	"net/http"

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

// RateLimitingTransport allows limiting max RPS
// at the transport layer
type RateLimitingTransport struct {
	Next http.RoundTripper
}

func (r *RateLimitingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return r.Next.RoundTrip(request)
}
