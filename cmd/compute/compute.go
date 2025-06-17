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

func init() {
	exocmd.RootCmd.AddCommand(ComputeCmd)
}
