package instance

import (
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/cmd/compute"
	"github.com/exoscale/cli/pkg/clicmd"
)

var instancePrivnetCmd = clicmd.Deprecate(&cobra.Command{
	Use:     "private-network",
	Short:   "Manage Compute instance Private Networks",
	Aliases: []string{"privnet"},
}, compute.PrivateNetworkDeprecation)

func init() {
	instanceCmd.AddCommand(instancePrivnetCmd)
}
