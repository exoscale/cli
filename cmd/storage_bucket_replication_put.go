package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/spf13/cobra"
)

func init() {
	storageBucketReplicationPutCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", zoneFlagMsg)
	storageBucketReplicationCmd.AddCommand(storageBucketReplicationPutCmd)
}

var storageBucketReplicationPutCmd = &cobra.Command{
	Use:   "put sos://BUCKET file://replication.json",
	Short: "Put replication configuration",
	Long: `Put a replication configuration for a bucket. Bucket versioning needs to be enabled
for both source and target bucket

Example of a valid replication configuration:
{
    "Role": "role-uuid",
    "Rules": [
        {
            "Status": "Enabled",
            "Priority": 1,
            "DeleteMarkerReplication": { "Status": "Disabled" },
            "Filter" : { "Prefix": ""},
            "Destination": {
                "Bucket": "target-bucket"
            },
            "ID": "foo"
        }
    ]
}

More information at https://docs.aws.amazon.com/cli/latest/reference/s3api/put-bucket-replication.html#options & https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketReplication.html`,
	Args: cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)
		args[1] = strings.TrimPrefix(args[1], "file://")

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		var configuration sos.BucketReplicationConf

		confFile, err := os.Open(args[1])
		if err != nil {
			return err
		}

		jsonParsr := json.NewDecoder(confFile)
		if err = jsonParsr.Decode(&configuration); err != nil {
			return err
		}

		bucket := args[0]

		zone, err := cmd.Flags().GetString(zoneFlagLong)
		if err != nil {
			return err
		}

		storage, err := sos.NewStorageClient(
			gContext,
			sos.ClientOptWithZone(zone),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		s3conf := configuration.ToS3()

		err = storage.PutBucketReplication(cmd.Context(), bucket, s3conf)

		return err
	},
}
