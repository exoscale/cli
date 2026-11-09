package storage

import (
	"fmt"
	"strings"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/storage/sos"
	"github.com/exoscale/cli/utils"
	"github.com/spf13/cobra"
)

func init() {
	storageBucketReplicationShowCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", zoneFlagMsg)
	storageBucketReplicationCmd.AddCommand(storageBucketReplicationShowCmd)
}

var storageBucketReplicationShowCmd = &cobra.Command{
	Use:   "show sos://BUCKET",
	Short: "Retrieve replication configuration",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {

		args[0] = strings.TrimPrefix(args[0], sos.BucketPrefix)

		clicmd.CmdSetZoneFlagFromDefault(cmd)
		return clicmd.CmdCheckRequiredFlags(cmd, []string{zoneFlagLong})
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		bucket := args[0]

		zone, err := cmd.Flags().GetString(zoneFlagLong)
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

		o, err := storage.GetBucketReplication(exocmd.GContext, bucket)
		if err != nil {
			return err
		}

		return utils.PrintOutput(o, nil)
	},
}
