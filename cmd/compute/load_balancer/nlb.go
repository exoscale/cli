package load_balancer

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

var nlbCmd = &cobra.Command{
	Use:     "load-balancer",
	Short:   "Network Load Balancers management",
	Aliases: []string{"nlb"},
}

func init() {
	compute.ComputeCmd.AddCommand(nlbCmd)
}
