package config

import (
	"encoding/json"
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
	Debug           bool
	FollowRedirects bool
	Version         string
	Cache           bool
	Insecure        bool
	MaxConnections  int
	Certificate     string
	PrivateKey      string
}

func (c *Config) String() string {
	s, _ := json.MarshalIndent(c, "", "\t")
	return string(s)
}
