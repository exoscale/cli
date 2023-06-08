package flags

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/storage/sos/object"
)

const (
	OlderThan          = "older-than"
	OlderThanTimestamp = "older-than-timestamp"
	NewerThan          = "newer-than"
	NewerThanTimestamp = "newer-than-timestamp"

	iso8601TimestampLayout = "2006-01-02T15:04:05Z07:00"
)

func parseTimestamp(s string) (time.Time, error) {
	return time.Parse(iso8601TimestampLayout, s)
}

func AddTimeFilterFlags(cmd *cobra.Command) {
	cmd.Flags().Duration(OlderThan, 0, "TODO")
	cmd.Flags().String(OlderThanTimestamp, "", "TODO")
	cmd.Flags().Duration(NewerThan, 0, "TODO")
	cmd.Flags().String(NewerThanTimestamp, "", "TODO")
}

func ValidateTimestampFlags(cmd *cobra.Command) error {
	flagsToValidate := []string{
		OlderThanTimestamp,
		NewerThanTimestamp,
	}

	for _, fl := range flagsToValidate {
		s, err := cmd.Flags().GetString(fl)
		if err != nil {
			return err
		}

		if s == "" {
			continue
		}

		if _, err := parseTimestamp(s); err != nil {
			return err
		}
	}

	return nil
}

func TranslateTimeFilterFlagsToFilterFuncs(cmd *cobra.Command) ([]object.ObjectFilterFunc, error) {
	var filters []object.ObjectFilterFunc

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

type filterCreationFunc func(*cobra.Command) (object.ObjectFilterFunc, error)

func olderThanDurationFilter(cmd *cobra.Command) (object.ObjectFilterFunc, error) {
	dur, err := cmd.Flags().GetDuration(OlderThan)
	if err != nil {
		return nil, err
	}

	if dur == 0 {
		return nil, nil
	}

	return func(obj object.ObjectInterface) bool {
		return obj.GetLastModified().Before(time.Now().Add(-dur))
	}, nil
}

func olderThanTimestampFilter(cmd *cobra.Command) (object.ObjectFilterFunc, error) {
	timestampStr, err := cmd.Flags().GetString(OlderThanTimestamp)
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

	return func(obj object.ObjectInterface) bool {
		return obj.GetLastModified().Before(timestamp)
	}, nil
}

func newerThanDurationFilter(cmd *cobra.Command) (object.ObjectFilterFunc, error) {
	dur, err := cmd.Flags().GetDuration(NewerThan)
	if err != nil {
		return nil, err
	}

	if dur == 0 {
		return nil, nil
	}

	return func(obj object.ObjectInterface) bool {
		return obj.GetLastModified().After(time.Now().Add(-dur))
	}, nil
}

func newerThanTimestampFilter(cmd *cobra.Command) (object.ObjectFilterFunc, error) {
	timestampStr, err := cmd.Flags().GetString(NewerThanTimestamp)
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

	return func(obj object.ObjectInterface) bool {
		return obj.GetLastModified().After(timestamp)
	}, nil
}
