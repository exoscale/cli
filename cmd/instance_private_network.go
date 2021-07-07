package cmd

import (
	"github.com/spf13/cobra"
)

var computeInstancePrivnetCmd = &cobra.Command{
	Use:     "private-network",
	Short:   "Manage Compute instance Private Networks",
	Aliases: []string{"privnet"},
}

func init() {
	computeInstanceCmd.AddCommand(computeInstancePrivnetCmd)
}
