package cmd

import (
	"github.com/exoscale/cli/cmd/compute"
	"github.com/spf13/cobra"
)

const defaultTemplateVisibility = "public"

var instanceTemplateCmd = &cobra.Command{
	Use:     "instance-template",
	Short:   "Compute instance templates management",
	Aliases: []string{"template"},
}

func init() {
	compute.ComputeCmd.AddCommand(instanceTemplateCmd)
}
