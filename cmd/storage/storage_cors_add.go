package storage

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

const (
	storageCORSAddCmdFlagAllowedOrigin = "allowed-origin"
	storageCORSAddCmdFlagAllowedMethod = "allowed-method"
	storageCORSAddCmdFlagAllowedHeader = "allowed-header"
)

// CORSRuleFromCmdFlags returns a non-nil pointer to a sos.CORSRule struct if at least
// one of the CORS-related command flags is set.
func CORSRuleFromCmdFlags(flags *pflag.FlagSet) *sos.CORSRule {
	var cors *sos.CORSRule

	flags.VisitAll(func(flag *pflag.Flag) {
		switch flag.Name {
		case storageCORSAddCmdFlagAllowedOrigin:
			if v, _ := flags.GetStringSlice(storageCORSAddCmdFlagAllowedOrigin); len(v) > 0 {
				if cors == nil {
					cors = &sos.CORSRule{}
				}

				cors.AllowedOrigins = v
			}

		case storageCORSAddCmdFlagAllowedMethod:
			if v, _ := flags.GetStringSlice(storageCORSAddCmdFlagAllowedMethod); len(v) > 0 {
				if cors == nil {
					cors = &sos.CORSRule{}
				}

				cors.AllowedMethods = v
			}

		case storageCORSAddCmdFlagAllowedHeader:
			if v, _ := flags.GetStringSlice(storageCORSAddCmdFlagAllowedHeader); len(v) > 0 {
				if cors == nil {
					cors = &sos.CORSRule{}
				}

				cors.AllowedHeaders = v
			}

		default:
			return
		}
	})

	return cors
}

var storageCORSAddCmd = &cobra.Command{
	Use:   "add sos://BUCKET",
	Short: "Add a CORS configuration rule to a bucket",
	Long: `This command adds a new rule to the current bucket CORS
configuration.

Example:

    exo storage cors add sos://my-bucket \
        --allowed-origin "https://my-website.net" \
        --allowed-method "*"
`,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			exocmd.CmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		return exocmd.CmdCheckRequiredFlags(cmd, []string{
			storageCORSAddCmdFlagAllowedOrigin,
			storageCORSAddCmdFlagAllowedMethod,
		})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		bucket := args[0]

		storage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		cors := CORSRuleFromCmdFlags(cmd.Flags())
		if err := storage.AddBucketCORSRule(exocmd.GContext, bucket, cors); err != nil {
			return fmt.Errorf("unable to add rule to the bucket CORS configuration: %w", err)
		}

		if !globalstate.Quiet {
			return utils.PrintOutput(storage.ShowBucket(exocmd.GContext, bucket))
		}

		return nil
	},
}

func init() {
	storageCORSAddCmd.Flags().StringSlice(storageCORSAddCmdFlagAllowedOrigin, nil,
		"allowed origin (can be repeated multiple times)")
	storageCORSAddCmd.Flags().StringSlice(storageCORSAddCmdFlagAllowedMethod, nil,
		"allowed method (can be repeated multiple times)")
	storageCORSAddCmd.Flags().StringSlice(storageCORSAddCmdFlagAllowedHeader, nil,
		"allowed header (can be repeated multiple times)")
	storageCORSCmd.AddCommand(storageCORSAddCmd)
}
