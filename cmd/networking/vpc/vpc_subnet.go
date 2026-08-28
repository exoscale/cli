package vpc

import (
	"github.com/spf13/cobra"
)

var vpcSubnetCmd = &cobra.Command{
	Use:   "subnet",
	Short: "Manage VPC Subnets",
}

func init() {
	Cmd.AddCommand(vpcSubnetCmd)
}
