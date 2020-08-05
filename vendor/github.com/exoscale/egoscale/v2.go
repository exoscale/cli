package egoscale

// optionalString returns the dereferenced string value of v if not nil, otherwise an empty string.
func optionalString(v *string) string {
	if v != nil {
		return *v
	}

	return ""
}

// optionalInt64 returns the dereferenced int64 value of v if not nil, otherwise 0.
func optionalInt64(v *int64) int64 {
	if v != nil {
		return *v
	}

	return 0
}
