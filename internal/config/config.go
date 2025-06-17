package config

import (
	"time"
)

// Config encapsulates the runtime configuration options
type Config struct {
	VersionSet  bool
	QuietSet    bool
	MaxRPS      int
	Concurrency int
	Duration    time.Duration
	Output      string
	Method      string
	Timeout     time.Duration
	HTTP2       bool
	Host        string
	UserAgent   string
	Endpoint    string
	BasicAuth   string
	Headers     []string
}

// New instantiates a new config and returns a ptr to the instance
func New() *Config {
	return &Config{}
}
