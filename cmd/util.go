package cmd

import (
	"fmt"
	"net"
	"strconv"
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

// defaultIP returns the IP as string if not nil, otherwise the default value specified.
func defaultIP(i *net.IP, def string) string {
	if i != nil {
		return i.String()
	}

	return def
}

// defaultInt64 returns the value of the int64 pointer b if not nil, otherwise the default value specified.
func defaultInt64(i *int64, def int64) int64 {
	if i != nil {
		return *i
	}

	return def
}

// nonEmptyStringPtr returns a non-nil pointer to s if the string is not empty, otherwise nil.
func nonEmptyStringPtr(s string) *string {
	if s != "" {
		return &s
	}

	return nil
}

func isEmptyStringPtr(s *string) bool {
	return s == nil || *s == ""
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

// versionMajor returns major part of a version number (given "x.y(.z)", returns "x").
// If the input version is not in semver format, returns 0.
func versionMajor(version string) uint32 {
	parts := strings.Split(version, ".")

	if len(parts) > 0 {
		v, e := strconv.ParseUint(parts[0], 10, 32)
		if e != nil {
			return 0
		}

		return uint32(v)
	}

	return 0
}

// versionMinor returns minor part of a version number (given "x.y(.z)", returns "y").
// If the input version is not in semver format, returns 0.
func versionMinor(version string) uint32 {
	parts := strings.Split(version, ".")

	if len(parts) > 1 {
		v, e := strconv.ParseUint(parts[1], 10, 32)
		if e != nil {
			return 0
		}

		return uint32(v)
	}

	return 0
}

// versionIsNewer returns true if new version has potential deprecation
func versionIsNewer(old, new string) bool {
	return (versionMajor(new) >= versionMajor(old)) ||
		(versionMajor(new) == versionMajor(old) && versionMinor(new) >= versionMinor(old))
}

// versionsAreEquivalent returns true if new and old versions both have same major and minor numbers
func versionsAreEquivalent(a, b string) bool {
	return (versionMajor(b) == versionMajor(a) && versionMinor(b) == versionMinor(a))
}
