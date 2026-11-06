package instance

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/cmd/compute"
)

var instancePrivnetCmd = &cobra.Command{
	Use:     "private-network",
	Short:   "Manage Compute instance Private Networks" + compute.PrivateNetworkDeprecatedShort,
	Long:    compute.PrivateNetworkDeprecationNotice,
	Aliases: []string{"privnet"},
}

func init() {
	instanceCmd.AddCommand(instancePrivnetCmd)
}
