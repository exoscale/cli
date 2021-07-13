package cmd

import (
	"github.com/spf13/cobra"
)

var computeInstanceTemplateCmd = &cobra.Command{
	Use:     "instance-template",
	Short:   "Compute instance templates management",
	Aliases: []string{"template"},
}

func init() {
	computeCmd.AddCommand(computeInstanceTemplateCmd)
}
