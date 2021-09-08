package cmd

import (
	"github.com/spf13/cobra"
)

var computeInstanceEIPCmd = &cobra.Command{
	Use:     "elastic-ip",
	Short:   "Manage Compute instance Elastic IP addresses",
	Aliases: []string{"eip"},
}

func init() {
	computeInstanceCmd.AddCommand(computeInstanceEIPCmd)
}
