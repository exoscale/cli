package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var storageHeaderAddCmd = &cobra.Command{
	Use:   "add sos://BUCKET/(OBJECT|PREFIX/)",
	Short: "Add HTTP headers to an object",
	Long: fmt.Sprintf(`This command adds response HTTP headers to objects.

Example:

    exo storage headers add sos://my-bucket/data.json \
        --content-type=application/json \
        --cache-control=no-store

Note: adding an already existing header will overwrite its value.

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

		if headers := storageHeadersFromCmdFlags(cmd.Flags()); headers == nil {
			cmdExitOnUsageError(cmd, "no header flag specified")
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			prefix string
		)

		recursive, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}

		parts := strings.SplitN(args[0], "/", 2)
		bucket, prefix = parts[0], parts[1]

		storage, err := newStorageClient(
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		headers := storageHeadersFromCmdFlags(cmd.Flags())
		if err := storage.updateObjectsHeaders(bucket, prefix, headers, recursive); err != nil {
			return fmt.Errorf("unable to add headers to object: %w", err)
		}

		if !gQuiet && !recursive && !strings.HasSuffix(prefix, "/") {
			return output(storage.showObject(bucket, prefix))
		}

		if !gQuiet {
			fmt.Println("Headers added successfully")
		}

		return nil
	},
}

func init() {
	storageHeaderAddCmd.Flags().BoolP("recursive", "r", false,
		"add headers recursively (with object prefix only)")
	storageHeaderAddCmd.Flags().String(strings.ToLower(storageObjectHeaderCacheControl), "",
		`value for "Cache-Control" header`)
	storageHeaderAddCmd.Flags().String(strings.ToLower(storageObjectHeaderContentDisposition), "",
		`value for "Content-Disposition" header`)
	storageHeaderAddCmd.Flags().String(strings.ToLower(storageObjectHeaderContentEncoding), "",
		`value for "Content-Encoding" header`)
	storageHeaderAddCmd.Flags().String(strings.ToLower(storageObjectHeaderContentLanguage), "",
		`value for "Content-Language" header`)
	storageHeaderAddCmd.Flags().String(strings.ToLower(storageObjectHeaderContentType), "",
		`value for "Content-Type" header`)
	storageHeaderAddCmd.Flags().String(strings.ToLower(storageObjectHeaderExpires), "",
		`value for "Expires" header`)
	storageHeaderCmd.AddCommand(storageHeaderAddCmd)
}

func (c *storageClient) updateObjectHeaders(bucket, key string, headers map[string]*string) error {
	object, err := c.copyObject(bucket, key)
	if err != nil {
		return err
	}

	lookupHeader := func(key string, fallback *string) *string {
		if v, ok := headers[key]; ok {
			return v
		}
		return fallback
	}

	object.CacheControl = lookupHeader(storageObjectHeaderCacheControl, object.CacheControl)
	object.ContentDisposition = lookupHeader(storageObjectHeaderContentDisposition, object.ContentDisposition)
	object.ContentEncoding = lookupHeader(storageObjectHeaderContentEncoding, object.ContentEncoding)
	object.ContentLanguage = lookupHeader(storageObjectHeaderContentLanguage, object.ContentLanguage)
	object.ContentType = lookupHeader(storageObjectHeaderContentType, object.ContentType)

	// For some reason, the AWS SDK doesn't use the same type for the "Expires"
	// header in GetObject (*string) and CopyObject (*time.Time)...
	if v, ok := headers[storageObjectHeaderExpires]; ok {
		t, err := time.Parse(time.RFC822, aws.ToString(v))
		if err != nil {
			return fmt.Errorf(`invalid "Expires" header value %q, expecting RFC822 format`, aws.ToString(v))
		}
		object.Expires = &t
	}

	_, err = c.CopyObject(gContext, object)
	return err
}

func (c *storageClient) updateObjectsHeaders(bucket, prefix string, headers map[string]*string, recursive bool) error {
	return c.forEachObject(bucket, prefix, recursive, func(o *s3types.Object) error {
		return c.updateObjectHeaders(bucket, aws.ToString(o.Key), headers)
	})
}

// storageHeadersFromCmdFlags returns a non-nil map if at least
// one of the header-related command flags is set.
func storageHeadersFromCmdFlags(flags *pflag.FlagSet) map[string]*string {
	var headers map[string]*string

	flags.VisitAll(func(flag *pflag.Flag) {
		switch flag.Name {
		case strings.ToLower(storageObjectHeaderCacheControl):
			if v := flag.Value.String(); v != "" {
				if headers == nil {
					headers = make(map[string]*string)
				}

				headers[storageObjectHeaderCacheControl] = aws.String(v)
			}

		case strings.ToLower(storageObjectHeaderContentDisposition):
			if v := flag.Value.String(); v != "" {
				if headers == nil {
					headers = make(map[string]*string)
				}

				headers[storageObjectHeaderContentDisposition] = aws.String(v)
			}

		case strings.ToLower(storageObjectHeaderContentEncoding):
			if v := flag.Value.String(); v != "" {
				if headers == nil {
					headers = make(map[string]*string)
				}

				headers[storageObjectHeaderContentEncoding] = aws.String(v)
			}

		case strings.ToLower(storageObjectHeaderContentLanguage):
			if v := flag.Value.String(); v != "" {
				if headers == nil {
					headers = make(map[string]*string)
				}

				headers[storageObjectHeaderContentLanguage] = aws.String(v)
			}

		case strings.ToLower(storageObjectHeaderContentType):
			if v := flag.Value.String(); v != "" {
				if headers == nil {
					headers = make(map[string]*string)
				}

				headers[storageObjectHeaderContentType] = aws.String(v)
			}

		case strings.ToLower(storageObjectHeaderExpires):
			if v := flag.Value.String(); v != "" {
				if headers == nil {
					headers = make(map[string]*string)
				}

				headers[storageObjectHeaderExpires] = aws.String(v)
			}

		default:
			return
		}
	})

	return headers
}
