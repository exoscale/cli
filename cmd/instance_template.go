package cmd

import (
	"github.com/spf13/cobra"
)

var instanceTemplateCmd = &cobra.Command{
	Use:     "instance-template",
	Short:   "Compute instance templates management",
	Aliases: []string{"template"},
}

func init() {
	computeCmd.AddCommand(instanceTemplateCmd)
}
