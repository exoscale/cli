package storage

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
)

var storageRbCmd = &cobra.Command{
	Use:   "rb sos://BUCKET",
	Short: "Delete a bucket",

	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			exocmd.CmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		bucket := args[0]

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			if !utils.AskQuestion(exocmd.GContext, fmt.Sprintf("Are you sure you want to delete %s%s?", sos.BucketPrefix, bucket)) {
				return nil
			}
		}

		storage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptZoneFromBucket(exocmd.GContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		if err := storage.DeleteBucket(exocmd.GContext, bucket, recursive); err != nil {
			return fmt.Errorf("unable to delete bucket: %w", err)
		}

		if !globalstate.Quiet {
			fmt.Printf("Bucket %s%s deleted successfully\n", sos.BucketPrefix, bucket)
		}

		return nil
	},
}

func init() {
	storageRbCmd.Flags().BoolP("force", "f", false, exocmd.CmdFlagForceHelp)
	storageRbCmd.Flags().BoolP("recursive", "r", false,
		"empty the bucket before deleting it")
	storageCmd.AddCommand(storageRbCmd)
}
