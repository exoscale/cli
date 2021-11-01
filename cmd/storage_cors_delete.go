package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

var storageCORSDeleteCmd = &cobra.Command{
	Use:     "delete sos://BUCKET",
	Aliases: []string{"del"},
	Short:   "Delete the CORS configuration of a bucket",

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], storageBucketPrefix)

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		bucket := args[0]

		certsFile, err := cmd.Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			if !askQuestion(fmt.Sprintf("Are you sure you want to delete bucket %s CORS configuration?",
				bucket)) {
				return nil
			}
		}

		storage, err := newStorageClient(
			storageClientOptWithCertsFile(certsFile),
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		if err := storage.deleteBucketCORS(bucket); err != nil {
			return fmt.Errorf("unable to delete bucket CORS configuration: %w", err)
		}

		if !gQuiet {
			fmt.Println("CORS configuration deleted successfully")
		}

		return nil
	},
}

func init() {
	storageCORSDeleteCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	storageCORSCmd.AddCommand(storageCORSDeleteCmd)
}

func (c *storageClient) deleteBucketCORS(bucket string) error {
	_, err := c.DeleteBucketCors(gContext, &s3.DeleteBucketCorsInput{Bucket: &bucket})
	return err
}
