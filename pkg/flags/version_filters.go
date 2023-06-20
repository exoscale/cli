package flags

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/storage/sos/object"
)

const (
	Versions        = "versions"
	OnlyVersions    = "only-versions"
	ExcludeVersions = "exclude-versions"
)

var (
	VersionRegex = regexp.MustCompile(`v?\d+`)
)

func AddVersionsFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(Versions, false, "list all versions of objects(if the bucket is versioned)")
	cmd.Flags().StringSlice(OnlyVersions, []string{}, "limit the versions to be listed; implies --"+Versions)
	cmd.Flags().StringSlice(ExcludeVersions, []string{}, "exclude versions from being listed; implies --"+Versions)
}

func validateVersions(versions []string) error {
	for _, v := range versions {
		if !VersionRegex.MatchString(v) {
			return fmt.Errorf("%q is not a valid version id(865029700534464769) or version number(v123)", v)
		}
	}

	return nil
}

func ValidateVersionFlags(cmd *cobra.Command) error {
	vsToInclude, err := cmd.Flags().GetStringSlice(OnlyVersions)
	if err != nil {
		return err
	}

	vsToExclude, err := cmd.Flags().GetStringSlice(ExcludeVersions)
	if err != nil {
		return err
	}

	if len(vsToInclude) > 0 && len(vsToExclude) > 0 {
		return fmt.Errorf("--%s and --%s are mutually exclusive", OnlyVersions, ExcludeVersions)
	}

	if err := validateVersions(vsToInclude); err != nil {
		return err
	}

	return validateVersions(vsToExclude)
}

func TranslateVersionFilterFlagsToFilterFuncs(cmd *cobra.Command) ([]object.ObjectVersionFilterFunc, error) {
	vsToInclude, err := cmd.Flags().GetStringSlice(OnlyVersions)
	if err != nil {
		return nil, err
	}

	if len(vsToInclude) > 0 {
		return []object.ObjectVersionFilterFunc{onlyVersionsFilter(vsToInclude)}, nil
	}

	vsToExclude, err := cmd.Flags().GetStringSlice(ExcludeVersions)
	if err != nil {
		return nil, err
	}

	if len(vsToExclude) > 0 {
		return []object.ObjectVersionFilterFunc{excludeVersionsFilter(vsToExclude)}, nil
	}

	return nil, nil
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

		if *ovi.GetVersionId() == version {
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
