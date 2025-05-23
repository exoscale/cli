package deploy_target

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

var deployTargetCmd = &cobra.Command{
	Use:     "deploy-target",
	Short:   "Compute instance Deploy Targets management",
	Aliases: []string{"dt"},
}

func init() {
	compute.ComputeCmd.AddCommand(deployTargetCmd)
}
