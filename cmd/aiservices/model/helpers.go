package model

// int64PtrIfNonZero returns a pointer to v if it's non-zero, otherwise nil.
func int64PtrIfNonZero(v int64) *int64 {
	if v == 0 {
		return nil
	}
	return &v
}
