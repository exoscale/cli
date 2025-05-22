package blockstorage

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

var blockstorageCmd = &cobra.Command{
	Use:     "block-storage",
	Short:   "Block Storage management",
	Aliases: []string{"block", "bs"},
}

func init() {
	compute.ComputeCmd.AddCommand(blockstorageCmd)
}
