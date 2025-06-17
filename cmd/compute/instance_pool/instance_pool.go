package instance_pool

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

var instancePoolCmd = &cobra.Command{
	Use:     "instance-pool",
	Short:   "Instance Pools management",
	Aliases: []string{"pool"},
}

func init() {
	compute.ComputeCmd.AddCommand(instancePoolCmd)
}
