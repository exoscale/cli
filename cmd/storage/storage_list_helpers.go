package storage

import (
	"fmt"
	"strconv"
	"time"

	"github.com/exoscale/cli/pkg/storage/sos/object"
)

// listTimeFilters returns the object filters matching the
// --older-than, --older-than-timestamp, --newer-than and --newer-than-timestamp
// values. A zero duration or a zero timestamp adds no filter.
func listTimeFilters(
	olderThan time.Duration,
	olderThanTimestamp time.Time,
	newerThan time.Duration,
	newerThanTimestamp time.Time,
) []object.ObjectFilterFunc {
	var filters []object.ObjectFilterFunc

	if olderThan != 0 {
		filters = append(filters, object.OlderThanFilterFunc(time.Now().Add(-olderThan)))
	}

	if !olderThanTimestamp.IsZero() {
		filters = append(filters, object.OlderThanFilterFunc(olderThanTimestamp))
	}

	if newerThan != 0 {
		filters = append(filters, object.NewerThanFilterFunc(time.Now().Add(-newerThan)))
	}

	if !newerThanTimestamp.IsZero() {
		filters = append(filters, object.NewerThanFilterFunc(newerThanTimestamp))
	}

	return filters
}

// listVersionFilters returns the version filter matching
// the --only-versions or --exclude-versions values, if any.
func listVersionFilters(onlyVersions, excludeVersions []string) []object.ObjectVersionFilterFunc {
	if len(onlyVersions) > 0 {
		return []object.ObjectVersionFilterFunc{onlyVersionsFilter(onlyVersions)}
	}

	if len(excludeVersions) > 0 {
		return []object.ObjectVersionFilterFunc{excludeVersionsFilter(excludeVersions)}
	}

	return nil
}

func doesVersionMatch(ovi object.ObjectVersionInterface, matchVersions []string) bool {
	for _, version := range matchVersions {
		if len(version) > 0 && version[0] == 'v' {
			vUint64, err := strconv.ParseUint(version[1:], 10, 64)
			if err != nil {
				fmt.Printf("invalid version number %q, cannot match\n", version)

				continue
			}

			if ovi.GetVersionNumber() == vUint64 {
				return true
			}
		}

		if *ovi.GetVersionID() == version {
			return true
		}
	}

	return false
}

func onlyVersionsFilter(acceptedVersions []string) object.ObjectVersionFilterFunc {
	return func(ovi object.ObjectVersionInterface) bool {
		return doesVersionMatch(ovi, acceptedVersions)
	}
}

func excludeVersionsFilter(vsToExclude []string) object.ObjectVersionFilterFunc {
	return func(ovi object.ObjectVersionInterface) bool {
		return !doesVersionMatch(ovi, vsToExclude)
	}
}
