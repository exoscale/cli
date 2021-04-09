package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
)

var storageMetadataAddCmd = &cobra.Command{
	Use:   "add sos://BUCKET/(OBJECT|PREFIX/) KEY=VALUE...",
	Short: "Add key/value metadata to an object",
	Long: fmt.Sprintf(`This command adds key/value metadata to an object.

Example:

    exo storage metadata add sos://my-bucket/object-a \
        k1=v1 \
        k2=v2

Note: adding an already existing key will overwrite its value.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&storageShowObjectOutput{}), ", ")),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], storageBucketPrefix)

		if !strings.Contains(args[0], "/") {
			cmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", args[0]))
		}

		for _, kv := range args[1:] {
			if !strings.Contains(kv, "=") {
				cmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", kv))
			}
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket   string
			prefix   string
			metadata = make(map[string]string)
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

		for _, kv := range args[1:] {
			parts = strings.Split(kv, "=")
			metadata[parts[0]] = parts[1]
		}

		storage, err := newStorageClient(
			storageClientOptWithCertsFile(certsFile),
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %v", err)
		}

		if err := storage.addObjectsMetadata(bucket, prefix, metadata, recursive); err != nil {
			return fmt.Errorf("unable to add metadata to object: %s", err)
		}

		if !gQuiet && !recursive && !strings.HasSuffix(prefix, "/") {
			return output(storage.showObject(bucket, prefix))
		}

		if !gQuiet {
			fmt.Println("Metadata added successfully")
		}

		return nil
	},
}

func init() {
	storageMetadataAddCmd.Flags().BoolP("recursive", "r", false,
		"add metadata recursively (with object prefix only)")
	storageMetadataCmd.AddCommand(storageMetadataAddCmd)
}

func (c *storageClient) addObjectMetadata(bucket, key string, metadata map[string]string) error {
	object, err := c.copyObject(bucket, key)
	if err != nil {
		return err
	}

	if len(object.Metadata) == 0 {
		object.Metadata = make(map[string]string)
	}

	for k, v := range metadata {
		object.Metadata[k] = v
	}

	_, err = c.CopyObject(gContext, object)
	return err
}

func (c *storageClient) addObjectsMetadata(bucket, prefix string, metadata map[string]string, recursive bool) error {
	return c.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.addObjectMetadata(bucket, aws.ToString(o.Key), metadata)
	})
}
