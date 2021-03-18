package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
)

var storageDeleteCmd = &cobra.Command{
	Use:     "delete <bucket>/[object | prefix/]",
	Aliases: []string{"rm"},
	Short:   "Delete objects",
	Long: `This command deletes objects stored in a bucket.

If you want to target objects under a "directory" prefix, suffix the path
argument with "/":

    exo storage delete my-bucket/
    exo storage delete -r my-bucket/some-directory/
`,

	PreRun: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		if !strings.Contains(args[0], "/") {
			args[0] = args[0] + "/"
		}
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string
		)

		parts := strings.SplitN(args[0], "/", 2)
		bucket = parts[0]
		if len(parts) > 1 {
			prefix = parts[1]

			// Special case: the caller wants to target objects at the root of
			// the bucket, in this case the prefix is empty so we set the key
			// to a symbolic value that shall be removed later on.
			if prefix == "" {
				prefix = "/"
			}
		}

		certsFile, err := cmd.Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			if !askQuestion(fmt.Sprintf("Are you sure you want to delete %s%s?", bucket, prefix)) {
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

		deleted, err := storage.deleteObjects(bucket, prefix, recursive)
		if err != nil {
			return fmt.Errorf("unable to delete objects: %s", err)
		}

		if verbose {
			for _, o := range deleted {
				fmt.Println(aws.ToString(o.Key))
			}
		}

		return nil
	},
}

func init() {
	storageDeleteCmd.Flags().BoolP("force", "f", false,
		"attempt to delete objects without prompting for confirmation")
	storageDeleteCmd.Flags().BoolP("recursive", "r", false, "delete objects recursively")
	storageDeleteCmd.Flags().BoolP("verbose", "v", false, "output deleted objects")
	storageCmd.AddCommand(storageDeleteCmd)
}

func (c *storageClient) deleteObjects(bucket, prefix string, recursive bool) ([]s3types.DeletedObject, error) {
	deleteList := make([]s3types.ObjectIdentifier, 0)
	err := c.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
		deleteList = append(deleteList, s3types.ObjectIdentifier{Key: o.Key})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error listing objects to delete: %s", err)
	}

	// The S3 DeleteObjects API call is limited to 1000 keys per call, as a
	// precaution we're batching deletes.
	maxKeys := 1000
	deleted := make([]s3types.DeletedObject, 0)

	for i := 0; i < len(deleteList); i += maxKeys {
		j := i + maxKeys
		if j > len(deleteList) {
			j = len(deleteList)
		}

		res, err := c.DeleteObjects(gContext, &s3.DeleteObjectsInput{
			Bucket: &bucket,
			Delete: &s3types.Delete{Objects: deleteList[i:j]},
		})
		if err != nil {
			return nil, err
		}

		deleted = append(deleted, res.Deleted...)
	}

	return deleted, nil
}
