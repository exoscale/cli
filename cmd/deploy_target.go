package cmd

import (
	"github.com/spf13/cobra"
)

var deployTargetCmd = &cobra.Command{
	Use:     "deploy-target",
	Short:   "Deploy Targets management",
	Aliases: []string{"dt"},
}

func init() {
	computeCmd.AddCommand(deployTargetCmd)
}
