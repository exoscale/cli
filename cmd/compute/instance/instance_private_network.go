package instance

import (
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/cmd/compute"
)

var instancePrivnetCmd = exocmd.Deprecate(&cobra.Command{
	Use:     "private-network",
	Short:   "Manage Compute instance Private Networks",
	Aliases: []string{"privnet"},
}, compute.PrivateNetworkDeprecation)

func init() {
	instanceCmd.AddCommand(instancePrivnetCmd)
}
