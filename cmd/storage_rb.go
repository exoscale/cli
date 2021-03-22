package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
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

		certsFile, err := cmd.Flags().GetString("certs-file")
		if err != nil {
			return err
		}

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
			storageClientOptWithCertsFile(certsFile),
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %s", err)
		}

		if err := storage.deleteBucket(bucket, recursive); err != nil {
			return fmt.Errorf("unable to delete bucket: %s", err)
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

func (c storageClient) deleteBucket(bucket string, recursive bool) error {
	if recursive {
		if _, err := c.deleteObjects(bucket, "", true); err != nil {
			return fmt.Errorf("error deleting objects: %s", err)
		}
	}

	// Delete dangling multipart uploads preventing bucket deletion.
	res, err := c.ListMultipartUploads(gContext, &s3.ListMultipartUploadsInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return fmt.Errorf("error listing dangling multipart uploads: %s", err)
	}
	for _, mp := range res.Uploads {
		if _, err = c.AbortMultipartUpload(gContext, &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(bucket),
			Key:      mp.Key,
			UploadId: mp.UploadId,
		}); err != nil {
			return fmt.Errorf("error aborting dangling multipart upload: %s", err)
		}
	}

	if _, err := c.DeleteBucket(gContext, &s3.DeleteBucketInput{Bucket: aws.String(bucket)}); err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "BucketNotEmpty" {
				return errors.New("bucket is not empty, either delete files before or use flag `-r`")
			}
		}

		return fmt.Errorf("unable to retrieve bucket CORS configuration: %s", err)
	}

	return nil
}
