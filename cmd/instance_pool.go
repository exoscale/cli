package cmd

import (
	"github.com/spf13/cobra"
)

var instancePoolCmd = &cobra.Command{
	Use:     "instance-pool",
	Short:   "Instance Pools management",
	Aliases: []string{"pool"},
}

func init() {
	computeCmd.AddCommand(instancePoolCmd)
}
