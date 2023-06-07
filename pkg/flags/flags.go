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

	iso8601TimestampLayout = "2006-01-02T15:04:05Z07:00"
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
	cmd.Flags().String(VersionsOlderThanTimestampFlag, "", "TODO")
	cmd.Flags().Duration(VersionsNewerThanFlag, 0, "TODO")
	cmd.Flags().String(VersionsNewerThanTimestampFlag, "", "TODO")
}

func parseTimestamp(s string) (time.Time, error) {
	return time.Parse(iso8601TimestampLayout, s)
}

func ValidateVersionsFlags(cmd *cobra.Command) error {
	flagsToValidate := []string{
		VersionsOlderThanTimestampFlag,
		VersionsNewerThanTimestampFlag,
	}

	for _, fl := range flagsToValidate {
		s, err := cmd.Flags().GetString(fl)
		if err != nil {
			return err
		}

		if _, err := parseTimestamp(s); err != nil {
			return err
		}
	}

	return nil
}

type filterCreationFunc func(*cobra.Command) (entities.ObjectVersionFilterFunc, error)

func TranslateFlagsToFilters(cmd *cobra.Command) ([]entities.ObjectVersionFilterFunc, error) {
	var filters []entities.ObjectVersionFilterFunc

	filterCreationFuncs := []filterCreationFunc{
		olderThanDurationFilter,
		olderThanTimestampFilter,
		newerThanDurationFilter,
		newerThanTimestampFilter,
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

func olderThanTimestampFilter(cmd *cobra.Command) (entities.ObjectVersionFilterFunc, error) {
	timestampStr, err := cmd.Flags().GetString(VersionsOlderThanTimestampFlag)
	if err != nil {
		return nil, err
	}

	if timestampStr == "" {
		return nil, nil
	}

	timestamp, err := parseTimestamp(timestampStr)
	if err != nil {
		return nil, err
	}

	return func(obj entities.ObjectVersionInterface) bool {
		return obj.GetLastModified().Before(timestamp)
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

func newerThanTimestampFilter(cmd *cobra.Command) (entities.ObjectVersionFilterFunc, error) {
	timestampStr, err := cmd.Flags().GetString(VersionsNewerThanTimestampFlag)
	if err != nil {
		return nil, err
	}

	if timestampStr == "" {
		return nil, nil
	}

	timestamp, err := parseTimestamp(timestampStr)
	if err != nil {
		return nil, err
	}

	return func(obj entities.ObjectVersionInterface) bool {
		return obj.GetLastModified().After(timestamp)
	}, nil
}
