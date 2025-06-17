package validation

import (
	"errors"
	"net/http"
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

// ParseHTTPHeaders takes an input of colon separated http
// headers and splits them into appropriate headers for a
// http.Request to utilise.
func ParseHTTPHeaders(input []string) http.Header {
	var headers http.Header
	for _, h := range input {
		split := strings.SplitN(h, ":", 2)
		if len(split) != 2 {
			// Throw away the bad header.
			// TODO: Might want to write to stderr tho?
			continue
		}
		k, v := split[0], split[1]
		if k == "" || v == "" {
			// Throw away the bad header
			// TODO: Might want to write to stderr tho?
			continue
		}
		headers.Add(k, v)
	}
	return headers
}
