package blockstorage

import (
	"github.com/spf13/cobra"
)

var blockstorageSnapshotCmd = &cobra.Command{
	Use:     "snapshot",
	Short:   "Block Storage Snapshot management",
	Aliases: []string{"snap", "shot"},
}

func init() {
	blockstorageCmd.AddCommand(blockstorageSnapshotCmd)
}
