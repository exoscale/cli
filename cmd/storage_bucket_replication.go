package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	storageBucketReplicationCmd.Flags().StringP(zoneFlagLong, zoneFlagShort, "", zoneFlagMsg)
	storageBucketCmd.AddCommand(storageBucketReplicationCmd)
}

var storageBucketReplicationCmd = &cobra.Command{
	Use:   "replication",
	Short: "Object Storage Bucket replication management",
}

var storageBucketReplicationCmdLongHelp = func() string {
	return "Object Storage Bucket replication management"
}
