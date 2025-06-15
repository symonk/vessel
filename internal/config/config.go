package config

import "time"

// Config houses application-wide configuration, user supplied through the
// command line
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
}

// New instantiates a new config and returns a ptr to the instance
func New() *Config {
	return &Config{}
}
