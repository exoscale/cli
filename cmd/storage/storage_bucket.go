package storage

import (
	"github.com/exoscale/cli/cmd/storage/lifecycle"
	"github.com/spf13/cobra"
)

func init() {
	storageCmd.AddCommand(storageBucketCmd)
	storageBucketCmd.AddCommand(lifecycle.Cmd)
}

var storageBucketCmd = &cobra.Command{
	Use:              "bucket",
	Aliases:          []string{"b"},
	Short:            "Object Storage Bucket management",
	TraverseChildren: true,
	Long:             storageBucketCmdLongHelp(),
}

var storageBucketCmdLongHelp = func() string {
	// TODO
	return "Object Storage Bucket management"
}
