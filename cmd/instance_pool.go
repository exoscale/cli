package cmd

import (
	"github.com/spf13/cobra"
)

var instancePoolCmd = &cobra.Command{
	Use:   "instancepool",
	Short: "Instance Pools management",
}

func init() {
	RootCmd.AddCommand(instancePoolCmd)
}
