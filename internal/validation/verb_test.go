package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethodsAreValidatedCorrectly(t *testing.T) {
	tests := map[string]struct {
		verb    string
		allowed bool
	}{
		"GET":     {"get", true},
		"POST":    {"post", true},
		"PUT":     {"put", true},
		"DELETE":  {"delete", true},
		"PATCH":   {"patch", true},
		"HEAD":    {"head", true},
		"OPTIONS": {"options", true},
		"CONNECT": {"connect", true},
		"TRACE":   {"trace", true},
		"NO":      {"no", false},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, Allowed.Permitted(test.verb), test.allowed)
		})
	}
}
