package cmd

import (
	"github.com/spf13/cobra"
)

var computeInstanceSnapshotCmd = &cobra.Command{
	Use:     "snapshot",
	Short:   "Manage Compute instance snapshots",
	Aliases: []string{"snap"},
}

func init() {
	computeInstanceCmd.AddCommand(computeInstanceSnapshotCmd)
}
