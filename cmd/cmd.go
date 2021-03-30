package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const cmdFlagForceHelp = "attempt to perform the operation without prompting for confirmation"

// cmdCheckRequiredFlags evaluates the specified flags as parsed in the cobra.Command flagset to check that
// their value is unset (i.e. null/empty/zero, depending on the type), and returns a multierror listing all
// flags missing a required value.
func cmdCheckRequiredFlags(cmd *cobra.Command, flags []string) error {
	var err *multierror.Error

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		for _, fn := range flags {
			if flag.Name == fn {
				var hasValue bool

				switch flag.Value.Type() {
				case "string", "stringSlice":
					if flag.Value.String() != "" {
						hasValue = true
					}

				case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "int64", "uint64",
					"float32", "float64":
					if flag.Value.String() != "0" {
						hasValue = true
					}
				}

				if !hasValue {
					err = multierror.Append(err, fmt.Errorf("no value specified for flag %q", fn))
				}
			}
		}
	})

	return err.ErrorOrNil()
}

// cmdSetZoneFlagFromDefault attempts to set the "--zone" flag value based on the current active account's
// default zone setting if set. This is a convenience helper, there is no guarantee that the flag will be
// set once this function returns.
func cmdSetZoneFlagFromDefault(cmd *cobra.Command) {
	if cmd.Flag("zone").Value.String() == "" {
		cmd.Flag("zone").Value.Set(gCurrentAccount.DefaultZone) // nolint:errcheck
	}
}

func cmdExitOnUsageError(cmd *cobra.Command, reason string) {
	cmd.PrintErrln(fmt.Sprintf("error: %s", reason))
	cmd.Usage() // nolint:errcheck
	os.Exit(1)
}

// completeVMNames is a Cobra Command.ValidArgsFunction that returns the list of Compute instance names belonging to
// the current user for shell auto-completion.
func completeVMNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	list, err := listVMs()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}
	return list.(*vmListOutput).names(), cobra.ShellCompDirectiveNoFileComp
}

func getCommaflag(p string) []string {
	if p == "" {
		return nil
	}

	p = strings.Trim(p, ",")
	args := strings.Split(p, ",")

	res := []string{}

	for _, arg := range args {
		if arg == "" {
			continue
		}
		res = append(res, strings.TrimSpace(arg))
	}

	return res
}
