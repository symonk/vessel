package validation

import (
	"errors"
	"strings"
)

// ParseBasicAuth attempts to parse the user defined basic auth credentials.
// an error is returned.
//
// ParseBasicAuth adheres to RFC 7617 and allows empty credentials
func ParseBasicAuth(input string) (user, pass string, err error) {
	if !strings.Contains(input, ":") {
		return "", "", errors.New("basic auth missing ':' separator")
	}
	parts := strings.SplitN(input, ":", 2)
	return parts[0], parts[1], nil
}
