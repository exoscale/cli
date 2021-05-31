package cmd

import (
	"github.com/spf13/cobra"
)

var deployTargetCmd = &cobra.Command{
	Use:     "deploytarget",
	Short:   "Deploy Targets management",
	Aliases: []string{"dt"},
}

func init() {
	vmCmd.AddCommand(deployTargetCmd)
}
