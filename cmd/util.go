package cmd

import (
	"fmt"
	"strings"
)

// isInList returns true if v exists in the specified list, false otherwise.
func isInList(list []string, v string) bool {
	for _, lv := range list {
		if lv == v {
			return true
		}
	}

	return false
}

// ellipString truncates the string s with an ellipsis character if longer
// than maxLen.
func ellipString(s string, maxLen int) string {
	ellipsis := "…"

	if len(s) > maxLen {
		return s[0:maxLen-1] + ellipsis
	}

	return s
}

// defaultString returns the value of the string pointer s if not nil, otherwise the default value specified.
func defaultString(s *string, def string) string {
	if s != nil {
		return *s
	}

	return def
}

// defaultBool returns the value of the bool pointer b if not nil, otherwise the default value specified.
func defaultBool(b *bool, def bool) bool {
	if b != nil {
		return *b
	}

	return def
}

// sliceToMap returns a map[string]string from a slice of KEY=VALUE formatted
// strings.
// This function is used to obtain a map[string]string from CLI flags, as the
// current CLI flags parsing module used (github.com/spf13/pflag) implements
// a "StringToString" type flag but doesn't support passing empty values,
// which we need in some cases (e.g. resetting labels).
func sliceToMap(v []string) (map[string]string, error) {
	m := make(map[string]string)

	for i := range v {
		parts := strings.SplitN(v[i], "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid value %q, expected format KEY=VALUE", v[i])
		}

		m[parts[0]] = parts[1]
	}

	return m, nil
}
