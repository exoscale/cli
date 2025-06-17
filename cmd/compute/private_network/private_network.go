package private_network

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

var privateNetworkCmd = &cobra.Command{
	Use:     "private-network",
	Short:   "Private Networks management",
	Aliases: []string{"privnet"},
}

func init() {
	compute.ComputeCmd.AddCommand(privateNetworkCmd)
}
