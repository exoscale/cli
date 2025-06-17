package security_group

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

var securityGroupCmd = &cobra.Command{
	Use:     "security-group",
	Short:   "Security Groups management",
	Aliases: []string{"sg"},
}

func init() {
	compute.ComputeCmd.AddCommand(securityGroupCmd)
}
