package trace

import "time"

type Trace struct {
	DnsStart     time.Time
	DnsDone      time.Duration
	ConnectStart time.Time
	ConnectDone  time.Duration
	TlsStart     time.Time
	TlsDone      time.Duration
}
