package requester

import (
	"net/http"

	"github.com/symonk/vessel/internal/config"
)

// GenerateTemplateRequest generates a template http request that can
// be cloned internally when sending > 1.  Care should be taken with the
// generated request bodies in terms of 're-reading' them etc.  This
// function is currently a naive implementation and offers no complexity
// for anything other than a simple GET request.
func GenerateTemplateRequest(cfg *config.Config) (*http.Request, error) {
	return http.NewRequest(cfg.Method, cfg.Endpoint, nil)
}
