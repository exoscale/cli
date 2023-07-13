package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/storage/sos"
)

var storagePresignCmd = &cobra.Command{
	Use:   "presign sos://BUCKET/OBJECT",
	Short: "Generate a pre-signed URL to an object",

	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			cmdExitOnUsageError(cmd, "invalid arguments")
		}

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

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
			gContext,
			sos.ClientOptZoneFromBucket(gContext, bucket),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		// TODO presign versions
		url, err := storage.GenPresignedURL(gContext, method, bucket, key, expires)
		if err != nil {
			return fmt.Errorf("unable to pre-sign %s%s/%s: %w", sos.BucketPrefix, bucket, key, err)
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
