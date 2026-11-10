package private_network

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/spf13/cobra"
)

var privateNetworkCmd = clicmd.Deprecate(&cobra.Command{
	Use:     "private-network",
	Short:   "Private Networks management",
	Aliases: []string{"privnet"},
}, compute.PrivateNetworkDeprecation)

func init() {
	compute.ComputeCmd.AddCommand(privateNetworkCmd)
}
