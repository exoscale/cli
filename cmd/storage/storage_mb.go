package storage

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

var storageMbCmd = &cobra.Command{
	Use:     "mb sos://BUCKET",
	Aliases: []string{"create"},
	Short:   "Create a new bucket",
	Long: fmt.Sprintf(`This command creates a new bucket.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&sos.ShowBucketOutput{}), ", ")),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			exocmd.CmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		exocmd.CmdSetZoneFlagFromDefault(cmd)

		return exocmd.CmdCheckRequiredFlags(cmd, []string{"zone"})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		bucket := args[0]

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		acl, err := cmd.Flags().GetString("acl")
		if err != nil {
			return err
		}

		storage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptWithZone(zone),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		if err := storage.CreateNewBucket(exocmd.GContext, bucket, acl); err != nil {
			return fmt.Errorf("unable to create bucket: %w", err)
		}

		if !globalstate.Quiet {
			return utils.PrintOutput(storage.ShowBucket(exocmd.GContext, bucket))
		}

		return nil
	},
}

func init() {
	storageMbCmd.Flags().String("acl", "",
		fmt.Sprintf("canned ACL to set on bucket (%s)", strings.Join(sos.BucketCannedACLToStrings(), "|")))
	storageMbCmd.Flags().StringP(exocmd.ZoneFlagLong, exocmd.ZoneFlagShort, "", exocmd.ZoneFlagMsg)
	storageCmd.AddCommand(storageMbCmd)
}
