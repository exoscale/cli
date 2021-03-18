package cmd

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

var storageCORSResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset the CORS configuration of a bucket",

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

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
			if !askQuestion(fmt.Sprintf("Are you sure you want to reset bucket %s CORS configuration?",
				bucket)) {
				return nil
			}
		}

		storage, err := newStorageClient(
			storageClientOptWithCertsFile(certsFile),
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %v", err)
		}

		if err := storage.resetBucketCORS(bucket); err != nil {
			return fmt.Errorf("unable to reset bucket CORS configuration: %s", err)
		}

		if !gQuiet {
			fmt.Println("CORS configuration reset successfully")
		}

		return nil
	},
}

func init() {
	storageCORSResetCmd.Flags().BoolP("force", "f", false,
		"attempt to reset CORS configuration without prompting for confirmation")
	storageCORSCmd.AddCommand(storageCORSResetCmd)
}

func (c *storageClient) resetBucketCORS(bucket string) error {
	_, err := c.DeleteBucketCors(gContext, &s3.DeleteBucketCorsInput{Bucket: &bucket})
	return err
}
