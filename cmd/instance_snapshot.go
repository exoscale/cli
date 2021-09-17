package cmd

import (
	"github.com/spf13/cobra"
)

var instanceSnapshotCmd = &cobra.Command{
	Use:     "snapshot",
	Short:   "Manage Compute instance snapshots",
	Aliases: []string{"snap"},
}

func init() {
	instanceCmd.AddCommand(instanceSnapshotCmd)
}
