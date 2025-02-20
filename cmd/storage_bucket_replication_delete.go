package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/spf13/cobra"
)

func init() {
	storageBucketReplicationDeleteCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", zoneFlagMsg)
	storageBucketReplicationCmd.AddCommand(storageBucketReplicationDeleteCmd)
}

var storageBucketReplicationDeleteCmd = &cobra.Command{
	Use:   "delete sos://BUCKET",
	Short: "Delete replication configuration",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
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

		err = storage.DeleteBucketReplication(cmd.Context(), bucket)
		return err
	},
}
