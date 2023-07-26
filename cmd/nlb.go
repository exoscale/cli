package cmd

import (
	"github.com/spf13/cobra"
)

var nlbCmd = &cobra.Command{
	Use:     "load-balancer",
	Short:   "Network Load Balancers management",
	Aliases: []string{"nlb"},
}

func init() {
	computeCmd.AddCommand(nlbCmd)
}
