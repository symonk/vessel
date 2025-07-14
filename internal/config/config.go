package config

import (
	"encoding/json"
	"net/http"
	"time"
)

// Config encapsulates the runtime configuration options
type Config struct {
	QuietSet        bool
	MaxRPS          int
	Concurrency     int
	Duration        time.Duration
	Method          string
	Timeout         time.Duration
	HTTP2           bool
	Host            string
	UserAgent       string
	Endpoint        string
	BasicAuth       string
	Headers         []string
	Amount          int64
	FollowRedirects bool
}

// New instantiates a new config and returns a ptr to the instance.
func New() *Config {
	return &Config{
		Concurrency: 10,
		Duration:    time.Second * 10,
		Method:      http.MethodGet,
	}
}

func (c *Config) String() string {
	s, _ := json.MarshalIndent(c, "", "\t")
	return string(s)
}
