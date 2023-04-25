package cmd

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

var storageHeaderCmd = &cobra.Command{
	Use:   "headers",
	Short: "Manage objects HTTP headers",
}

func init() {
	storageCmd.AddCommand(storageHeaderCmd)
}

// sos.ObjectHeadersFromS3 returns mutable object headers in a human-friendly
// key/value form.
func sos.ObjectHeadersFromS3(o *s3.GetObjectOutput) map[string]string {
	headers := make(map[string]string)

	if o.CacheControl != nil {
		headers[sos.ObjectHeaderCacheControl] = aws.ToString(o.CacheControl)
	}
	if o.ContentDisposition != nil {
		headers[sos.ObjectHeaderContentDisposition] = aws.ToString(o.ContentDisposition)
	}
	if o.ContentEncoding != nil {
		headers[sos.ObjectHeaderContentEncoding] = aws.ToString(o.ContentEncoding)
	}
	if o.ContentLanguage != nil {
		headers[sos.ObjectHeaderContentLanguage] = aws.ToString(o.ContentLanguage)
	}
	if o.ContentType != nil {
		headers[sos.ObjectHeaderContentType] = aws.ToString(o.ContentType)
	}
	if o.Expires != nil {
		headers[sos.ObjectHeaderExpires] = o.Expires.String()
	}

	return headers
}
