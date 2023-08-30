package v3

// String returns the value of the string pointer p if not nil, otherwise empty string.
func String(p *string) string {
	if p != nil {
		return *p
	}

	return ""
}

// DefaultString returns the value of the string pointer p if not nil, otherwise the default value specified.
func DefaultString(p *string, d string) string {
	if p != nil {
		return *p
	}

	return d
}

// FromString returns a non-nil pointer to string s.
func FromString(s string) *string {
	return &s
}

// Bool returns the value of the bool pointer p if not nil, otherwise returns false.
func Bool(p *bool) bool {
	if p != nil {
		return *p
	}

	return false
}

// DefaultBool returns the value of the bool pointer p if not nil, otherwise the default value specified.
func DefaultBool(p *bool, d bool) bool {
	if p != nil {
		return *p
	}

	return d
}

// FromBool returns a non-nil pointer to bool b.
func FromBool(b bool) *bool {
	return &b
}

// Int64 returns the value of the int64 pointer p if not nil, otherwise returns 0.
func Int64(p *int64) int64 {
	if p != nil {
		return *p
	}

	return 0
}

// DefaultInt64 returns the value of the int64 pointer p if not nil, otherwise the default value specified.
func DefaultInt64(p *int64, d int64) int64 {
	if p != nil {
		return *p
	}

	return d
}

// FromInt64 returns a non-nil pointer to int64 b.
func FromInt64(i int64) *int64 {
	return &i
}

// StringSlice returns the value of string slice pointer p if not nil, otherwise returns nil.
func StringSlice(p *[]string) []string {
	if p != nil {
		return *p
	}

	return nil
}

// FromStringSlice returns a non-nil pointer to string slice s.
func FromStringSlice(s []string) *[]string {
	if s != nil {
		return &s
	}

	return nil
}
