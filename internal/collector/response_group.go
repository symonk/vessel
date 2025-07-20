package collector

import (
	"fmt"
	"sync"
)

// StatusCodeCounter keeps track of how many of each type of
// response status code has been seen.
type StatusCodeCounter struct {
	mu sync.Mutex
	m  map[int]int
}

func NewStatusCodeCounter() *StatusCodeCounter {
	return &StatusCodeCounter{
		m: make(map[int]int),
	}
}

// Increments increments the counter for a particular status codes.
//
// This is safe for concurrent use
func (s *StatusCodeCounter) Increment(code int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[code] = s.m[code] + 1
}

// String returns a string representation of the captures response
// codes.
// TODO: can use strings.Builder
func (s *StatusCodeCounter) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	str := "Breakdown\n"
	for k, v := range s.m {
		str += fmt.Sprintf("\t[%d]: %d", k, v)
		str += "\n"
	}
	return str
}

// Count returns the total number of responses that have been
// recorded by the code grouper instance
func (s *StatusCodeCounter) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	t := 0
	for _, v := range s.m {
		t += v
	}
	return t
}
