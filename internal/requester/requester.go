package requester

import (
	"net/http"
	"time"

	"github.com/symonk/vessel/internal/collector"
)

type Requester interface {
	Go()
	Wait()
}

// Requester takes a request and fans out many instances
// of that request until either the maximum count is reached
// or the duration has been surpassed.
type RequestSender struct {
	collector collector.Collector
	client    http.Client
}

// New instantiates a new instance of Requester and returns
// the ptr to it.
func New(collector collector.Collector, timeout time.Duration) *RequestSender {
	return &RequestSender{
		collector: collector,
		client: http.Client{
			Timeout: timeout,
		},
	}
}

func (u *RequestSender) Go()   {}
func (u *RequestSender) Wait() {}
