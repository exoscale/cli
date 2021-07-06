package cmd

import (
	"github.com/spf13/cobra"
)

var computeInstanceSGCmd = &cobra.Command{
	Use:     "security-group",
	Short:   "Manage Compute instance Security Groups",
	Aliases: []string{"sg"},
}

func init() {
	computeInstanceCmd.AddCommand(computeInstanceSGCmd)
}
