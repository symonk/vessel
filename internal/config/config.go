package config

import "time"

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
}
