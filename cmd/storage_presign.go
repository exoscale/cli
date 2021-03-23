package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/cobra"
)

var storagePresignCmd = &cobra.Command{
	Use:   "presign sos://BUCKET/OBJECT",
	Short: "Generate a pre-signed URL to an object",

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], storageBucketPrefix)

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			bucket string
			key    string
		)

		certsFile, err := cmd.Flags().GetString("certs-file")
		if err != nil {
			return err
		}

		expires, err := cmd.Flags().GetDuration("expires")
		if err != nil {
			return err
		}

		method, err := cmd.Flags().GetString("method")
		if err != nil {
			return err
		}

		parts := strings.SplitN(args[0], "/", 2)
		bucket, key = parts[0], parts[1]

		storage, err := newStorageClient(
			storageClientOptWithCertsFile(certsFile),
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %v", err)
		}

		url, err := storage.genPresignedURL(method, bucket, key, expires)
		if err != nil {
			return fmt.Errorf("unable to pre-sign %s%s/%s: %s", storageBucketPrefix, bucket, key, err)
		}

		fmt.Println(url)

		return nil
	},
}

func init() {
	storagePresignCmd.Flags().StringP("method", "m", "get", "pre-signed URL method (get|put)")
	storagePresignCmd.Flags().DurationP("expires", "e", 900*time.Second,
		`expiration duration for the generated pre-signed URL (e.g. "1h45m", "30s"); supported units: "s", "m", "h"`)
	storageCmd.AddCommand(storagePresignCmd)
}

func (c *storageClient) genPresignedURL(method, bucket, key string, expires time.Duration) (string, error) {
	var (
		psURL *v4.PresignedHTTPRequest
		err   error
	)

	psClient := s3.NewPresignClient(c.Client, func(o *s3.PresignOptions) {
		if expires > 0 {
			o.Expires = expires
		}
	})

	switch method {
	case "get":
		psURL, err = psClient.PresignGetObject(gContext, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

	case "put":
		psURL, err = psClient.PresignPutObject(gContext, &s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})

	default:
		err = fmt.Errorf("unsupported method %q", method)
	}

	if err != nil {
		return "", err
	}

	return psURL.URL, nil
}
