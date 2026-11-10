package compute

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

var ComputeCmd = &cobra.Command{
	Use:        "compute",
	Short:      "Compute services management",
	Aliases:    []string{"c"},
	SuggestFor: []string{"vm", "aag", "firewall", "instancepool", "nlb"},
}

// PrivateNetworkDeprecation is the deprecation message of Private Network commands.
const PrivateNetworkDeprecation = `Private Networks are deprecated, use VPCs instead (see "exo networking vpc --help")`

func init() {
	exocmd.RootCmd.AddCommand(ComputeCmd)
}
