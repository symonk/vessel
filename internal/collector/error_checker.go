package collector

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
)

type ErrorType = string

const (
	// TODO: Build out later
	// TODO: Document in CLI --help indepth
	DNS        ErrorType = "DNS"
	TLS        ErrorType = "TLS"
	Timeout    ErrorType = "Timeout"
	Cancelled  ErrorType = "Cancelled"
	Connection ErrorType = "Connection"
	Unknown    ErrorType = "Unknown"
)

// ErrorMap defines a custom type of error types where the value is
// the number of times an error in that category has been seen.
type ErrorMap map[ErrorType]*atomic.Int64

// ErrorGrouper captures context on HTTP related errors in a synchronised fashion
// to better give context on the type (and number of) errors that fell into various
// buckets or groups.  Outputting all the errors throughout a run is verbose, grouping
// them into useful 'buckets' provides a better user experience.
//
// This is not exhaustive right now and will be growing as development continues.
//
// ErrorGrouper is synchronised internally and is safe for parallel use.
type ErrorGrouper struct {
	store ErrorMap
}

// NewErrGrouper instantiates a new ErrorGrouper and returns a ptr
// to the instance.
func NewErrGrouper() *ErrorGrouper {
	categories := map[ErrorType]*atomic.Int64{
		DNS:        new(atomic.Int64),
		TLS:        new(atomic.Int64),
		Timeout:    new(atomic.Int64),
		Cancelled:  new(atomic.Int64),
		Connection: new(atomic.Int64),
		Unknown:    new(atomic.Int64),
	}
	return &ErrorGrouper{
		store: categories,
	}
}

// Record captures and categorises an error
func (e *ErrorGrouper) Record(err error) {
	if e == nil || err == nil {
		return
	}
	switch {
	case errors.Is(err, context.Canceled):
		// The process was likely interrupted by a sigterm.  Requests in flight
		// or potentially ones that have not yet been sent will quickly fill
		// up this bucket.  This is NOT a case where the server was slow to respond,
		// that difference is important in reporting.
		e.store[Cancelled].Add(1)
	case errors.Is(err, context.DeadlineExceeded):
		// The time specified for a particular request likely was exceeded.  The
		// server is more than likely failing to respond within the clients expectations.
		e.store[Timeout].Add(1)

	case strings.Contains(err.Error(), "connection refused"), strings.Contains(err.Error(), "connection reset by peer"):
		// connection refused: TCP connection failures, nothing listening.
		// connection reset by peer: TCP connect success, server crashed/closed/rejected/RST.
		e.store[Connection].Add(1)
	// TODO: Reading required on the *net.OpError and various other potential concrete error
	// types exposed by Go's stdlib.
	default:
		e.store[Unknown].Add(1)
	}

}

// String implements fmt.Stringer and provides a useful summary of the
// errors received throughout vessells lifecycle which is used to compose
// the core result summary later.
func (e *ErrorGrouper) String() string {
	timedOut := e.store[Timeout].Load()
	cancelled := e.store[Cancelled].Load()
	unknown := e.store[Unknown].Load()
	connection := e.store[Connection].Load()
	all := timedOut + cancelled + unknown + connection
	return fmt.Sprintf("Total: %d: Timeout(%d), Cancelled(%d), Connection(%d), Unknown(%d)", all, timedOut, cancelled, connection, unknown)
}
