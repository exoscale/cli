package cmd

import (
	"github.com/spf13/cobra"
)

var deployTargetCmd = &cobra.Command{
	Use:     "deploy-target",
	Short:   "Compute instance Deploy Targets management",
	Aliases: []string{"dt"},
}

func init() {
	computeCmd.AddCommand(deployTargetCmd)
}
