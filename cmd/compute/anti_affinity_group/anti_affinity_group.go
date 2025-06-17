package anti_affinity_group

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

var antiAffinityGroupCmd = &cobra.Command{
	Use:     "anti-affinity-group",
	Short:   "Anti-Affinity Groups management",
	Aliases: []string{"aag"},
}

func init() {
	compute.ComputeCmd.AddCommand(antiAffinityGroupCmd)
}
