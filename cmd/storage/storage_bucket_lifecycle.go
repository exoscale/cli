package storage

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

func init() {
	storageBucketLifecycleCmd.Flags().StringP(exocmd.ZoneFlagLong, exocmd.ZoneFlagShort, "", exocmd.ZoneFlagMsg)
	storageBucketCmd.AddCommand(storageBucketLifecycleCmd)
}

var storageBucketLifecycleCmd = &cobra.Command{
	Use:   "lifecycle",
	Short: "Object Storage Bucket lifecycle management",
	Long:  storageBucketLifecycleCmdLongHelp(),
}

var storageBucketLifecycleCmdLongHelp = func() string {
	return "Object Storage Bucket lifecycle management"
}
