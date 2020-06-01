package cmd

import "github.com/spf13/cobra"

var nlbCmd = &cobra.Command{
	Use:   "nlb",
	Short: "Network Load Balancers management",
}

func init() {
	RootCmd.AddCommand(nlbCmd)
}
