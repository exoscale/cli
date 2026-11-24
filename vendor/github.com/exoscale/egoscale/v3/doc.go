// Package v3 is the Go client of the Exoscale API.
//
// Create a [Client] with [NewClient] and credentials from the
// [github.com/exoscale/egoscale/v3/credentials] package. Each API operation is
// a method of [Client], and its request and response types are generated from
// the Exoscale OpenAPI spec.
//
// API errors wrap a sentinel error of their HTTP status, such as [ErrNotFound],
// to be checked with [errors.Is].
package v3
