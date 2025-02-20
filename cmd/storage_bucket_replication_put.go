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
	Use:   "put sos://BUCKET file://configuration.json",
	Short: "Put replication configuration",
	Args:  cobra.ExactArgs(2),
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
