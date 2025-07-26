package stats

import "time"

type ReusedState = int64

const (
	NotReused ReusedState = 0
	WasReused ReusedState = 1
)

// Stats encapsulates the metrics of concern retrieved from a response.
// Stats serves as an intermediate type between the workers and the
// collector implementation.
type Stats struct {
	Err           error
	Latency       time.Duration
	StatusCode    int
	TimeOnDns     time.Duration
	TimeOnTls     time.Duration
	TimeOnConnect time.Duration
	TimeOnConn    time.Duration
	BytesSent     int64
	BytesReceived int64
	ReusedConn    ReusedState
}
