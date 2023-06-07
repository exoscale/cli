package flags

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/entities"
)

const (
	VersionsFlag                   = "versions"
	VersionsOlderThanFlag          = "versions-older-than"
	VersionsOlderThanTimestampFlag = "versions-older-than-timestamp"
	VersionsNewerThanFlag          = "versions-newer-than"
	VersionsNewerThanTimestampFlag = "versions-newer-than-timestamp"
	ExcludeCurrentVersionFlag      = "exclude-current-version"
)

var FlagToFilter = map[string]entities.ObjectFilterFunc{
	VersionsFlag:                   entities.AcceptAll,
	VersionsOlderThanFlag:          entities.AcceptAll,
	VersionsOlderThanTimestampFlag: entities.AcceptAll,
	VersionsNewerThanFlag:          entities.AcceptAll,
	VersionsNewerThanTimestampFlag: entities.AcceptAll,
	ExcludeCurrentVersionFlag:      entities.AcceptAll,
}

func AddVersionsFlags(cmd *cobra.Command) {
	cmd.Flags().Duration(VersionsOlderThanFlag, 0, "TODO")
	cmd.Flags().Duration(VersionsNewerThanFlag, 0, "TODO")
}

type filterCreationFunc func(*cobra.Command) (entities.ObjectVersionFilterFunc, error)

func TranslateFlagsToFilters(cmd *cobra.Command) ([]entities.ObjectVersionFilterFunc, error) {
	var filters []entities.ObjectVersionFilterFunc

	filterCreationFuncs := []filterCreationFunc{
		olderThanDurationFilter,
		newerThanDurationFilter,
	}

	for _, fcf := range filterCreationFuncs {
		newFilter, err := fcf(cmd)
		if err != nil {
			return nil, err
		}

		if newFilter != nil {
			filters = append(filters, newFilter)
		}
	}

	return filters, nil
}

func olderThanDurationFilter(cmd *cobra.Command) (entities.ObjectVersionFilterFunc, error) {
	dur, err := cmd.Flags().GetDuration(VersionsOlderThanFlag)
	if err != nil {
		return nil, err
	}

	if dur == 0 {
		return nil, nil
	}

	return func(obj entities.ObjectVersionInterface) bool {
		return obj.GetLastModified().Before(time.Now().Add(-dur))
	}, nil
}

func newerThanDurationFilter(cmd *cobra.Command) (entities.ObjectVersionFilterFunc, error) {
	dur, err := cmd.Flags().GetDuration(VersionsNewerThanFlag)
	if err != nil {
		return nil, err
	}

	if dur == 0 {
		return nil, nil
	}

	return func(obj entities.ObjectVersionInterface) bool {
		return obj.GetLastModified().After(time.Now().Add(-dur))
	}, nil
}
