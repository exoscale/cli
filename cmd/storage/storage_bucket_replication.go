package storage

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

func init() {
	storageBucketReplicationCmd.Flags().StringP(exocmd.ZoneFlagLong, exocmd.ZoneFlagShort, "", exocmd.ZoneFlagMsg)
	storageBucketCmd.AddCommand(storageBucketReplicationCmd)
}

var storageBucketReplicationCmd = &cobra.Command{
	Use:   "replication",
	Short: "Object Storage Bucket replication management",
	Long:  storageBucketReplicationCmdLongHelp(),
}

var storageBucketReplicationCmdLongHelp = func() string {
	return "Object Storage Bucket replication management"
}
