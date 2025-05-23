package deploy_target

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

var deployTargetCmd = &cobra.Command{
	Use:     "deploy-target",
	Short:   "Compute instance Deploy Targets management",
	Aliases: []string{"dt"},
}

func init() {
	exocmd.ComputeCmd.AddCommand(deployTargetCmd)
}
