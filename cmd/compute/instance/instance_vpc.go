package instance

import (
	"github.com/spf13/cobra"
)

var instanceVPCCmd = &cobra.Command{
	Use:   "vpc",
	Short: "Manage Compute instance VPC Subnet attachments",
}

func init() {
	instanceCmd.AddCommand(instanceVPCCmd)
}
