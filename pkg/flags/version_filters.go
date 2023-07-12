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
	versionFlagDoc  = "accepts comma separated version IDs(865029700534464769) and numbers(v123)"
)

var (
	VersionRegex       = regexp.MustCompile(`v?\d+`)
	VersionNumberRegex = regexp.MustCompile(`v\d+`)
)

func AddVersionsFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(Versions, false, "list all versions of objects(if the bucket is versioned)")
	cmd.Flags().StringSlice(OnlyVersions, []string{}, "limit the versions to be listed; "+versionFlagDoc+"; implies --"+Versions)
	cmd.Flags().StringSlice(ExcludeVersions, []string{}, "exclude versions from being listed; "+versionFlagDoc+"; implies --"+Versions)
}

func validateVersions(versions []string, stream bool) error {
	for _, v := range versions {
		if !VersionRegex.MatchString(v) {
			return fmt.Errorf("%q is not a valid version id(865029700534464769) or version number(v123)", v)
		}

		if stream && VersionNumberRegex.MatchString(v) {
			return fmt.Errorf("cannot use version number filter %q in combination with --stream flag", v)
		}
	}

	return nil
}

func ValidateVersionFlags(cmd *cobra.Command, useStream bool) error {
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

	stream := false
	if useStream {
		stream, err = cmd.Flags().GetBool("stream")
		if err != nil {
			return err
		}
	}

	if err := validateVersions(vsToInclude, stream); err != nil {
		return err
	}

	return validateVersions(vsToExclude, stream)
}

func TranslateVersionFilterFlagsToFilterFuncs(cmd *cobra.Command) ([]object.ObjectVersionFilterFunc, error) {
	var filters []object.ObjectVersionFilterFunc
	vsToInclude, err := cmd.Flags().GetStringSlice(OnlyVersions)
	if err != nil {
		return nil, err
	}

	if len(vsToInclude) > 0 {
		return append(filters, onlyVersionsFilter(vsToInclude)), nil
	}

	vsToExclude, err := cmd.Flags().GetStringSlice(ExcludeVersions)
	if err != nil {
		return nil, err
	}

	if len(vsToExclude) > 0 {
		return append(filters, excludeVersionsFilter(vsToExclude)), nil
	}

	return filters, nil
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
