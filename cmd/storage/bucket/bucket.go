package bucket

import (
	"github.com/exoscale/cli/cmd/storage/bucket/replication"
	"github.com/exoscale/cli/cmd/storage/lifecycle"
	"github.com/spf13/cobra"
)

func init() {
	Cmd.AddCommand(lifecycle.Cmd)
	Cmd.AddCommand(replication.Cmd)
}

var Cmd = &cobra.Command{
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
