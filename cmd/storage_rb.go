package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var storageRbCmd = &cobra.Command{
	Use:   "rb sos://BUCKET",
	Short: "Delete a bucket",

	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], storageBucketPrefix)
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
			if !askQuestion(fmt.Sprintf("Are you sure you want to delete %s%s?", storageBucketPrefix, bucket)) {
				return nil
			}
		}

		storage, err := newStorageClient(
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		if err := storage.DeleteBucket(bucket, recursive); err != nil {
			return fmt.Errorf("unable to delete bucket: %w", err)
		}

		if !gQuiet {
			fmt.Printf("Bucket %s%s deleted successfully\n", storageBucketPrefix, bucket)
		}

		return nil
	},
}

func init() {
	storageRbCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	storageRbCmd.Flags().BoolP("recursive", "r", false,
		"empty the bucket before deleting it")
	storageCmd.AddCommand(storageRbCmd)
}
