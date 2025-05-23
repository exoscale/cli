package anti_affinity_group

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

var antiAffinityGroupCmd = &cobra.Command{
	Use:     "anti-affinity-group",
	Short:   "Anti-Affinity Groups management",
	Aliases: []string{"aag"},
}

func init() {
	exocmd.ComputeCmd.AddCommand(antiAffinityGroupCmd)
}
