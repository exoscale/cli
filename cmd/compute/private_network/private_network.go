package private_network

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

var privateNetworkCmd = &cobra.Command{
	Use:     "private-network",
	Short:   "Private Networks management",
	Aliases: []string{"privnet"},
}

func init() {
	exocmd.ComputeCmd.AddCommand(privateNetworkCmd)
}
