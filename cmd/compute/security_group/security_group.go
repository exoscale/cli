package security_group

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

var securityGroupCmd = &cobra.Command{
	Use:     "security-group",
	Short:   "Security Groups management",
	Aliases: []string{"sg"},
}

func init() {
	exocmd.ComputeCmd.AddCommand(securityGroupCmd)
}
