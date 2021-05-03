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
