package cmd

import (
	"github.com/spf13/cobra"
)

var computeCmd = &cobra.Command{
	Use:        "compute",
	Short:      "Compute services management",
	Aliases:    []string{"c"},
	SuggestFor: []string{"vm", "aag", "firewall", "instancepool", "nlb"},
}

func init() {
	RootCmd.AddCommand(computeCmd)
}
