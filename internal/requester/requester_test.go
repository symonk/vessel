package requester

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/symonk/vessel/internal/collector"
	"github.com/symonk/vessel/internal/config"
	"github.com/symonk/vessel/internal/test/mockserver"
)

// TODO: Verify no response leaks at scale (i.e not closing bodies cleanly)

const (
	TenThousand int64 = 10_000
)

func TestRequestorForFixedAmountBehavesCorrectly(t *testing.T) {
	server := mockserver.New(mockserver.WithStatusCodeTestHandler())
	defer server.Close()
	cfg := &config.Config{
		Method:   http.MethodGet,
		Endpoint: fmt.Sprintf("%s/status/200", server.Server.URL),
		Amount:   TenThousand,
	}
	writer := new(bytes.Buffer)
	collector := collector.New(writer, cfg)
	req, err := GenerateTemplateRequest(cfg)
	assert.NoError(t, err)
	r := New(context.Background(), cfg, collector, req)
	// Waiting for the instance to not block is sufficient
	r.Wait()
	assert.Equal(t, server.Seen.Load(), int64(TenThousand))
}

func TestMultipleWorkersFixedAountBehavesCorrectly(t *testing.T) {
	server := mockserver.New(mockserver.WithStatusCodeTestHandler())
	defer server.Close()
	cfg := &config.Config{
		Method:      http.MethodGet,
		Endpoint:    fmt.Sprintf("%s/status/200", server.Server.URL),
		Amount:      TenThousand,
		Concurrency: runtime.NumCPU(),
	}
	writer := new(bytes.Buffer)
	collector := collector.New(writer, cfg)
	req, err := GenerateTemplateRequest(cfg)
	assert.NoError(t, err)
	r := New(context.Background(), cfg, collector, req)
	r.Wait()
	assert.Equal(t, server.Seen.Load(), int64(TenThousand))
}
