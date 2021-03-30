package cmd

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
)

var storageHeaderDeleteCmd = &cobra.Command{
	Use:     "delete sos://BUCKET/(OBJECT|PREFIX/)",
	Aliases: []string{"del"},
	Short:   "Delete HTTP headers from an object",
	Long: fmt.Sprintf(`This command deletes response HTTP headers from objects.

Example:

    exo storage headers delete sos://my-bucket/data.json \
        --cache-control \
        --expires

Note: the "Content-Type" header cannot be removed, it is reset to its default
value "application/binary".

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&storageShowObjectOutput{}), ", ")),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], storageBucketPrefix)

		if !strings.Contains(args[0], "/") {
			cmdExitOnUsageError(cmd, fmt.Sprintf("invalid argument: %q", args[0]))
		}

		var hasHeaderFlagsSet bool
		for _, flag := range []string{
			storageObjectHeaderCacheControl,
			storageObjectHeaderContentDisposition,
			storageObjectHeaderContentEncoding,
			storageObjectHeaderContentLanguage,
			storageObjectHeaderContentType,
			storageObjectHeaderExpires,
		} {
			if cmd.Flags().Changed(strings.ToLower(flag)) {
				hasHeaderFlagsSet = true
				break
			}
		}
		if !hasHeaderFlagsSet {
			cmdExitOnUsageError(cmd, "no header flag specified")
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket  string
			prefix  string
			headers []string
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

		storage, err := newStorageClient(
			storageClientOptWithCertsFile(certsFile),
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %v", err)
		}

		for _, header := range []string{
			storageObjectHeaderCacheControl,
			storageObjectHeaderContentDisposition,
			storageObjectHeaderContentEncoding,
			storageObjectHeaderContentLanguage,
			storageObjectHeaderContentType,
			storageObjectHeaderExpires,
		} {
			if ok, _ := cmd.Flags().GetBool(strings.ToLower(header)); ok {
				headers = append(headers, header)
			}
		}

		if err := storage.deleteObjectsHeaders(bucket, prefix, headers, recursive); err != nil {
			return fmt.Errorf("unable to add headers to object: %s", err)
		}

		if !gQuiet && !recursive && !strings.HasSuffix(prefix, "/") {
			return output(storage.showObject(bucket, prefix))
		}

		if !gQuiet {
			fmt.Println("Headers deleted successfully")
		}

		return nil
	},
}

func init() {
	storageHeaderDeleteCmd.Flags().BoolP("recursive", "r", false,
		"delete headers recursively (with object prefix only)")
	storageHeaderDeleteCmd.Flags().Bool(strings.ToLower(storageObjectHeaderCacheControl), false,
		`delete the "Cache-Control" header`)
	storageHeaderDeleteCmd.Flags().Bool(strings.ToLower(storageObjectHeaderContentDisposition), false,
		`delete the "Content-Disposition" header`)
	storageHeaderDeleteCmd.Flags().Bool(strings.ToLower(storageObjectHeaderContentEncoding), false,
		`delete the "Content-Encoding" header`)
	storageHeaderDeleteCmd.Flags().Bool(strings.ToLower(storageObjectHeaderContentLanguage), false,
		`delete the "Content-Language" header`)
	storageHeaderDeleteCmd.Flags().Bool(strings.ToLower(storageObjectHeaderContentType), false,
		`delete the "Content-Type" header`)
	storageHeaderDeleteCmd.Flags().Bool(strings.ToLower(storageObjectHeaderExpires), false,
		`delete the "Expires" header`)
	storageHeaderCmd.AddCommand(storageHeaderDeleteCmd)
}

func (c *storageClient) deleteObjectHeaders(bucket, key string, headers []string) error {
	object, err := c.copyObject(bucket, key)
	if err != nil {
		return err
	}

	for _, header := range headers {
		switch header {
		case storageObjectHeaderCacheControl:
			object.CacheControl = nil

		case storageObjectHeaderContentDisposition:
			object.ContentDisposition = nil

		case storageObjectHeaderContentEncoding:
			object.ContentEncoding = nil

		case storageObjectHeaderContentLanguage:
			object.ContentLanguage = nil

		case storageObjectHeaderContentType:
			object.ContentType = aws.String("application/binary")

		case storageObjectHeaderExpires:
			object.Expires = nil
		}
	}

	_, err = c.CopyObject(gContext, object)
	return err
}

func (c *storageClient) deleteObjectsHeaders(bucket, prefix string, headers []string, recursive bool) error {
	return c.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.deleteObjectHeaders(bucket, aws.ToString(o.Key), headers)
	})
}
