package cmd

import (
	"github.com/spf13/cobra"
)

var instanceSGCmd = &cobra.Command{
	Use:     "security-group",
	Short:   "Manage Compute instance Security Groups",
	Aliases: []string{"sg"},
}

func init() {
	instanceCmd.AddCommand(instanceSGCmd)
}
