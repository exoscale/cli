package cmd

import (
	"github.com/hashicorp/go-multierror"
)

// forEachZone executes the function f for each specified zone, and return a multierror.Error containing all
// errors that may have occurred during execution.
func forEachZone(zones []string, f func(zone string) error) error {
	meg := new(multierror.Group)

	for _, zone := range zones {
		zone := zone
		meg.Go(func() error {
			return f(zone)
		})
	}

	return meg.Wait().ErrorOrNil()
}
