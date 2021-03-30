package cmd

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

const (
	storageObjectHeaderCacheControl       = "Cache-Control"
	storageObjectHeaderContentDisposition = "Content-Disposition"
	storageObjectHeaderContentEncoding    = "Content-Encoding"
	storageObjectHeaderContentLanguage    = "Content-Language"
	storageObjectHeaderContentType        = "Content-Type"
	storageObjectHeaderExpires            = "Expires"
)

var storageHeaderCmd = &cobra.Command{
	Use:   "headers",
	Short: "Manage objects HTTP headers",
}

func init() {
	storageCmd.AddCommand(storageHeaderCmd)
}

// storageObjectHeadersFromS3 returns mutable object headers in a human-friendly
// key/value form.
func storageObjectHeadersFromS3(o *s3.GetObjectOutput) map[string]string {
	headers := make(map[string]string)

	if o.CacheControl != nil {
		headers[storageObjectHeaderCacheControl] = aws.ToString(o.CacheControl)
	}
	if o.ContentDisposition != nil {
		headers[storageObjectHeaderContentDisposition] = aws.ToString(o.ContentDisposition)
	}
	if o.ContentEncoding != nil {
		headers[storageObjectHeaderContentEncoding] = aws.ToString(o.ContentEncoding)
	}
	if o.ContentLanguage != nil {
		headers[storageObjectHeaderContentLanguage] = aws.ToString(o.ContentLanguage)
	}
	if o.ContentType != nil {
		headers[storageObjectHeaderContentType] = aws.ToString(o.ContentType)
	}
	if o.Expires != nil {
		headers[storageObjectHeaderExpires] = o.Expires.String()
	}

	return headers
}
