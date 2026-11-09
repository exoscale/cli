package storage

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/storage/sos"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
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
	PreRunE: func(c *cobra.Command, args []string) error {

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		clicmd.CmdSetZoneFlagFromDefault(c)
		return clicmd.CmdCheckRequiredFlags(c, []string{zoneFlagLong})
	},
	RunE: func(c *cobra.Command, args []string) error {
		bucket := args[0]

		zone, err := c.Flags().GetString(zoneFlagLong)
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
