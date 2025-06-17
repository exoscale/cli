package storage

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/storage/sos"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

func init() {
	storageBucketReplicationDeleteCmd.Flags().StringP(exocmd.ZoneFlagLong, exocmd.ZoneFlagShort, "", exocmd.ZoneFlagMsg)
	storageBucketReplicationCmd.AddCommand(storageBucketReplicationDeleteCmd)
}

var storageBucketReplicationDeleteCmd = &cobra.Command{
	Use:   "delete sos://BUCKET",
	Short: "Delete replication configuration",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(c *cobra.Command, args []string) error {

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		exocmd.CmdSetZoneFlagFromDefault(c)
		return exocmd.CmdCheckRequiredFlags(c, []string{exocmd.ZoneFlagLong})
	},
	RunE: func(c *cobra.Command, args []string) error {
		bucket := args[0]

		zone, err := c.Flags().GetString(exocmd.ZoneFlagLong)
		if err != nil {
			return err
		}

		storage, err := sos.NewStorageClient(
			exocmd.GContext,
			sos.ClientOptWithZone(zone),
		)
		if err != nil {
			return fmt.Errorf("unable to initialize storage client: %w", err)
		}

		err = storage.DeleteBucketReplication(exocmd.GContext, bucket)
		return err
	},
}
