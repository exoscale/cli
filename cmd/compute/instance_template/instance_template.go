package cmd

import (
	exocmd "github.com/exoscale/cli/cmd"
	"github.com/spf13/cobra"
)

var instanceTemplateCmd = &cobra.Command{
	Use:     "instance-template",
	Short:   "Compute instance templates management",
	Aliases: []string{"template"},
}

func init() {
	exocmd.ComputeCmd.AddCommand(instanceTemplateCmd)
}
