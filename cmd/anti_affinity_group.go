package cmd

import (
	"github.com/spf13/cobra"
)

var antiAffinityGroupCmd = &cobra.Command{
	Use:     "anti-affinity-group",
	Short:   "Compute instance Anti-Affinity Groups management",
	Aliases: []string{"aag"},
}

func init() {
	computeCmd.AddCommand(antiAffinityGroupCmd)
}
