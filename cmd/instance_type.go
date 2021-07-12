package cmd

import (
	"github.com/spf13/cobra"
)

var computeInstanceTypeCmd = &cobra.Command{
	Use:   "instance-type",
	Short: "Compute instance types management",
}

func init() {
	computeCmd.AddCommand(computeInstanceTypeCmd)
}
