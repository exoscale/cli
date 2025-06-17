package storage

import (
	"github.com/spf13/cobra"
)

func init() {
	storageCmd.AddCommand(storageBucketCmd)
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
