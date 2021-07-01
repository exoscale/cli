package cmd

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
