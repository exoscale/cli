package vpc

import (
	"github.com/spf13/cobra"
)

var vpcRouteCmd = &cobra.Command{
	Use:   "route",
	Short: "Manage VPC routes",
}

func init() {
	Cmd.AddCommand(vpcRouteCmd)
}
