package utils

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
)

// LoggableContentTypes holds a list of Request/Response content types that debug logs will print.
var LoggableContentTypes = []string{
	"application/json",
	"text/plain",
}

// PrintHeaders merges all headers in a single line for use in logs.
func PrintHeaders(h http.Header) string {
	out := make([]string, 0, len(h)*2)
	for k, v := range h {
		out = append(out, fmt.Sprintf("%s:%s", k, strings.Join(v, ",")))
	}
	sort.Strings(out)

	return strings.Join(out, " ")
}

// IsLoggableContentType returns true if headers contain printable 'Content-Type'.
// Function returns false on invalid input, make sure input is validated before calling it.
//
// NOTE: Only single "Content-Type" is valid per RFC 7231.
func IsLoggableContentType(h http.Header) bool {
	if len(h) == 0 || len(h.Values("Content-Type")) != 1 {
		return false
	}

	ct := strings.ToLower(h.Get("Content-Type"))
	for _, target := range LoggableContentTypes {
		if ct == target || strings.HasPrefix(ct, target+";") {
			return true
		}
	}

	return false
}
