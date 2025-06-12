package verb

import (
	"net/http"
	"strings"
)

type AllowedVerbs map[string]struct{}

var Allowed = AllowedVerbs{
	http.MethodGet:     {},
	http.MethodConnect: {},
	http.MethodDelete:  {},
	http.MethodHead:    {},
	http.MethodOptions: {},
	http.MethodPatch:   {},
	http.MethodPost:    {},
	http.MethodPut:     {},
	http.MethodTrace:   {},
}

// Permitted reports whether the provided verb is allowed.
// It normalizes the verb to uppercase before checking.
func (a AllowedVerbs) Permitted(verb string) bool {
	_, ok := a[strings.ToUpper(verb)]
	return ok
}
