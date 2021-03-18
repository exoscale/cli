package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
)

var storageMbCmd = &cobra.Command{
	Use:     "mb <name>",
	Aliases: []string{"create"},
	Short:   "Create a new bucket",
	Long: fmt.Sprintf(`This command creates a new bucket.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&storageShowBucketOutput{}), ", ")),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		cmdSetZoneFlagFromDefault(cmd)

		return cmdCheckRequiredFlags(cmd, []string{"zone"})
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		bucket := args[0]

		certsFile, err := cmd.Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		zone, err := cmd.Flags().GetString("zone")
		if err != nil {
			return err
		}

		acl, err := cmd.Flags().GetString("acl")
		if err != nil {
			return err
		}

		storage, err := newStorageClient(
			storageClientOptWithCertsFile(certsFile),
			storageClientOptWithZone(zone),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %s", err)
		}

		if err := storage.createBucket(bucket, acl); err != nil {
			return fmt.Errorf("unable to create bucket: %s", err)
		}

		if !gQuiet {
			return output(storage.showBucket(bucket))
		}

		return nil
	},
}

func init() {
	storageMbCmd.Flags().String("acl", "",
		fmt.Sprintf("canned ACL to set on bucket (%s)", strings.Join(s3BucketCannedACLToStrings(), "|")))
	storageMbCmd.Flags().StringP("zone", "z", "", "bucket zone")
	storageCmd.AddCommand(storageMbCmd)
}

func (c *storageClient) createBucket(name, acl string) error {
	s3Bucket := s3.CreateBucketInput{Bucket: aws.String(name)}

	if acl != "" {
		if !isInList(s3BucketCannedACLToStrings(), acl) {
			return fmt.Errorf("invalid canned ACL %q, supported values are: %s",
				acl,
				strings.Join(s3BucketCannedACLToStrings(), ", "))
		}

		s3Bucket.ACL = s3types.BucketCannedACL(acl)
	}

	_, err := c.CreateBucket(gContext, &s3Bucket)
	return err
}
