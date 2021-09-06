package cmd

import (
	"github.com/spf13/cobra"
)

var securityGroupCmd = &cobra.Command{
	Use:     "security-group",
	Short:   "Compute instance Security Groups management",
	Aliases: []string{"sg"},
}

func init() {
	computeCmd.AddCommand(securityGroupCmd)
}
