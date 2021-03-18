package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
)

var storageMetadataDeleteCmd = &cobra.Command{
	Use:     "delete <bucket>/<object | prefix/> <key> [key ...]",
	Aliases: []string{"del"},
	Short:   "Delete metadata from an object",
	Long: fmt.Sprintf(`This command deletes key/value metadata from an object.

Example:

    exo storage metadata delete my-bucket/object-a k1

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&storageShowObjectOutput{}), ", ")),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		if !strings.Contains(args[0], "/") {
			cmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", args[0]))
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string
		)

		certsFile, err := cmd.Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		parts := strings.SplitN(args[0], "/", 2)
		bucket, prefix = parts[0], parts[1]
		mdKeys := args[1:]

		storage, err := newStorageClient(
			storageClientOptWithCertsFile(certsFile),
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %v", err)
		}

		if err := storage.deleteObjectsMetadata(bucket, prefix, mdKeys, recursive); err != nil {
			return fmt.Errorf("unable to delete metadata from object: %s", err)
		}

		if !gQuiet && !recursive && !strings.HasSuffix(prefix, "/") {
			return output(storage.showObject(bucket, prefix))
		}

		if !gQuiet {
			fmt.Println("Metadata deleted successfully")
		}

		return nil
	},
}

func init() {
	storageMetadataDeleteCmd.Flags().BoolP("recursive", "r", false,
		"delete metadata recursively (with object prefix only)")
	storageMetadataCmd.AddCommand(storageMetadataDeleteCmd)
}

func (c *storageClient) deleteObjectMetadata(bucket, key string, mdKeys []string) error {
	object, err := c.copyObject(bucket, key)
	if err != nil {
		return err
	}

	for _, k := range mdKeys {
		if _, ok := object.Metadata[k]; !ok {
			return fmt.Errorf("key %q not found in current metadata", k)
		}
		delete(object.Metadata, k)
	}

	_, err = c.CopyObject(gContext, object)
	return err
}

func (c *storageClient) deleteObjectsMetadata(bucket, prefix string, mdKeys []string, recursive bool) error {
	return c.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.deleteObjectMetadata(bucket, aws.ToString(o.Key), mdKeys)
	})
}
