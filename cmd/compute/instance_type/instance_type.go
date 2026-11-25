package cmd

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"

	v3 "github.com/exoscale/egoscale/v3"
)

var instanceTypeCmd = &cobra.Command{
	Use:   "instance-type",
	Short: "Compute instance types management",
}

func init() {
	compute.ComputeCmd.AddCommand(instanceTypeCmd)
}

func zoneNames(zones []v3.ZoneName) []string {
	out := make([]string, 0, len(zones))
	for _, z := range zones {
		out = append(out, string(z))
	}
	return out
}
