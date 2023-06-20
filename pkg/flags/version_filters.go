package flags

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/storage/sos/object"
)

const (
	Versions        = "versions"
	OnlyVersions    = "only-versions"
	ExcludeVersions = "exclude-versions"
)

func AddVersionsFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(Versions, false, "list all versions of objects(if the bucket is versioned)")
	cmd.Flags().StringArray(OnlyVersions, []string{}, "limit the versions to be listed; implies --"+Versions)
	cmd.Flags().StringArray(ExcludeVersions, []string{}, "exclude versions from being listed; implies --"+Versions)
}

func ValidateVersionFlags(cmd *cobra.Command) error {
	vsToInclude, err := cmd.Flags().GetStringArray(OnlyVersions)
	if err != nil {
		return err
	}

	vsToExclude, err := cmd.Flags().GetStringArray(ExcludeVersions)
	if err != nil {
		return err
	}

	if len(vsToInclude) > 0 && len(vsToExclude) > 0 {
		return fmt.Errorf("--%s and --%s are mutually exclusive", OnlyVersions, ExcludeVersions)
	}

	return nil
}

func TranslateVersionFilterFlagsToFilterFuncs(cmd *cobra.Command) ([]object.ObjectVersionFilterFunc, error) {
	vsToInclude, err := cmd.Flags().GetStringArray(OnlyVersions)
	if err != nil {
		return nil, err
	}

	if len(vsToInclude) > 0 {
		return []object.ObjectVersionFilterFunc{onlyVersionsFilter(vsToInclude)}, nil
	}

	vsToExclude, err := cmd.Flags().GetStringArray(ExcludeVersions)
	if err != nil {
		return nil, err
	}

	if len(vsToExclude) > 0 {
		return []object.ObjectVersionFilterFunc{excludeVersionsFilter(vsToExclude)}, nil
	}

	return nil, nil
}

func onlyVersionsFilter(acceptedVersions []string) object.ObjectVersionFilterFunc {
	return func(ovi object.ObjectVersionInterface) bool {
		for _, v := range acceptedVersions {
			if *ovi.GetVersionId() == v {
				return true
			}
		}

		return false
	}
}

func excludeVersionsFilter(vsToExclude []string) object.ObjectVersionFilterFunc {
	return func(ovi object.ObjectVersionInterface) bool {
		for _, v := range vsToExclude {
			if *ovi.GetVersionId() == v {
				return false
			}
		}

		return true
	}
}
