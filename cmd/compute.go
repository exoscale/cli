package cmd

import (
	"github.com/spf13/cobra"
)

var computeCmd = &cobra.Command{
	Use:     "compute",
	Short:   "Compute services management",
	Aliases: []string{"c"},
}

func init() {
	RootCmd.AddCommand(computeCmd)
}
