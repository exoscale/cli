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

// PrivateNetworkDeprecatedShort is appended to the short help of Private Network commands.
const PrivateNetworkDeprecatedShort = " (deprecated, use VPC instead)"

// PrivateNetworkDeprecationNotice is prepended to the long help of Private Network commands.
const PrivateNetworkDeprecationNotice = `DEPRECATED: Private Networks are deprecated. Please use VPCs instead,
see "exo networking vpc --help".`

func init() {
	exocmd.RootCmd.AddCommand(ComputeCmd)
}
