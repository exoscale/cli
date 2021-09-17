package cmd

import (
	"github.com/spf13/cobra"
)

var instanceTypeCmd = &cobra.Command{
	Use:   "instance-type",
	Short: "Compute instance types management",
}

func init() {
	computeCmd.AddCommand(instanceTypeCmd)
}
