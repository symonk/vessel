package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsingBasicAuth(t *testing.T) {
	tests := map[string]struct {
		input      string
		expectFunc func(t *testing.T, user, pass string, err error)
	}{
		"correct": {
			input: "foo:bar",
			expectFunc: func(t *testing.T, user, pass string, err error) {
				assert.Equal(t, user, "foo")
				assert.Equal(t, pass, "bar")
				assert.NoError(t, err)
			},
		},
		"no_colon": {
			input: "userpw",
			expectFunc: func(t *testing.T, user, pass string, err error) {
				assert.Equal(t, user, "")
				assert.Equal(t, pass, "")
				assert.ErrorContains(t, err, "basic auth missing ':' separator")
			},
		},
		"empty_user": {
			input: ":pw",
			expectFunc: func(t *testing.T, user, pass string, err error) {
				assert.Equal(t, user, "")
				assert.Equal(t, pass, "pw")
				assert.NoError(t, err)
			},
		},
		"empty_pw": {
			input: "user:",
			expectFunc: func(t *testing.T, user, pass string, err error) {
				assert.Equal(t, user, "user")
				assert.Equal(t, pass, "")
				assert.NoError(t, err)
			},
		},
		"empty_both": {
			input: ":",
			expectFunc: func(t *testing.T, user, pass string, err error) {
				assert.Equal(t, user, "")
				assert.Equal(t, pass, "")
				assert.NoError(t, err)
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			u, p, err := ParseBasicAuth(test.input)
			test.expectFunc(t, u, p, err)
		})
	}
}
