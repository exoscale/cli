package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/exoscale/cli/pkg/storage/sos"
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

		storage, err := sos.NewStorageClient(
			storageClientOptZoneFromBucket(bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		url, err := storage.GenPresignedURL(method, bucket, key, expires)
		if err != nil {
			return fmt.Errorf("unable to pre-sign %s%s/%s: %w", storageBucketPrefix, bucket, key, err)
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
