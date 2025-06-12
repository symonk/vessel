package requester

import "github.com/symonk/vessel/internal/collector"

type Requester interface {
	Go()
	Wait()
}

// Requester takes a request and fans out many instances
// of that request until either the maximum count is reached
// or the duration has been surpassed.
type UnboundRequester struct {
	collector collector.Collector
}

// New instantiates a new instance of Requester and returns
// the ptr to it.
func New(collector collector.Collector) *UnboundRequester {
	return &UnboundRequester{
		collector: collector,
	}
}

func (u *UnboundRequester) Go()   {}
func (u *UnboundRequester) Wait() {}
