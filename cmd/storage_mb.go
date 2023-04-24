package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/spf13/cobra"
)

var storageMbCmd = &cobra.Command{
	Use:     "mb sos://BUCKET",
	Aliases: []string{"create"},
	Short:   "Create a new bucket",
	Long: fmt.Sprintf(`This command creates a new bucket.

Supported output template annotations: %s`,
		strings.Join(output.output.OutputterTemplateAnnotations(&storageShowBucketOutput{}), ", ")),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], storageBucketPrefix)

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
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
			storageClientOptWithZone(zone),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		if err := storage.CreateBucket(bucket, acl); err != nil {
			return fmt.Errorf("unable to create bucket: %w", err)
		}

		if !gQuiet {
			return printOutput(storage.ShowBucket(bucket))
		}

		return nil
	},
}

func init() {
	storageMbCmd.Flags().String("acl", "",
		fmt.Sprintf("canned ACL to set on bucket (%s)", strings.Join(s3BucketCannedACLToStrings(), "|")))
	storageMbCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", zoneFlagMsg)
	storageCmd.AddCommand(storageMbCmd)
}
