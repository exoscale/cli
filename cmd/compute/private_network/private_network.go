package private_network

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

var privateNetworkCmd = exocmd.Deprecate(&cobra.Command{
	Use:     "private-network",
	Short:   "Private Networks management",
	Aliases: []string{"privnet"},
}, compute.PrivateNetworkDeprecation)

func init() {
	compute.ComputeCmd.AddCommand(privateNetworkCmd)
}
