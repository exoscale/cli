package egoscale

import "errors"

// ErrNotFound represents an error indicating a non-existent resource.
var ErrNotFound = errors.New("resource not found")

// ErrTooManyFound represents an error indicating multiple results found for a single resource.
var ErrTooManyFound = errors.New("multiple resources found")
