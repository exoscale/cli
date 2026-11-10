package instance_pool

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

const (
	defaultInstanceType       = "standard.medium"
	defaultTemplateVisibility = "public"
)

var instancePoolCmd = &cobra.Command{
	Use:     "instance-pool",
	Short:   "Instance Pools management",
	Aliases: []string{"pool"},
}

func init() {
	compute.ComputeCmd.AddCommand(instancePoolCmd)
}
