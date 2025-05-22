package blockstorage

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

var blockstorageCmd = &cobra.Command{
	Use:     "block-storage",
	Short:   "Block Storage management",
	Aliases: []string{"block", "bs"},
}

func init() {
	exocmd.ComputeCmd.AddCommand(blockstorageCmd)
}
