package cmd

import (
	"github.com/spf13/cobra"
)

var blockstorageCmd = &cobra.Command{
	Use:     "block-storage",
	Short:   "Block Storage management",
	Aliases: []string{"block", "bs"},
}

func init() {
	computeCmd.AddCommand(blockstorageCmd)
}
