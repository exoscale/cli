package cmd

import (
	"github.com/spf13/cobra"
)

// snapshotCmd represents the snapshot command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Snapshots allow you to save the volume of your machine in its current state",
}

func init() {
	RootCmd.AddCommand(snapshotCmd)
}
