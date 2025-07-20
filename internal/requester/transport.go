package requester

import (
	"net/http"
)

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
