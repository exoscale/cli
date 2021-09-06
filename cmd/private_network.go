package cmd

import (
	"github.com/spf13/cobra"
)

var privateNetworkCmd = &cobra.Command{
	Use:     "private-network",
	Short:   "Compute instance Private Networks management",
	Aliases: []string{"privnet"},
}

func init() {
	computeCmd.AddCommand(privateNetworkCmd)
}
