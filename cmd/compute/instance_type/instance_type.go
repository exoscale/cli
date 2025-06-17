package cmd

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

var instanceTypeCmd = &cobra.Command{
	Use:   "instance-type",
	Short: "Compute instance types management",
}

func init() {
	compute.ComputeCmd.AddCommand(instanceTypeCmd)
}
