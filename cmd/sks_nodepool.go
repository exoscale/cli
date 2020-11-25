package cmd

import "github.com/spf13/cobra"

var sksNodepoolCmd = &cobra.Command{
	Use:     "nodepool",
	Short:   "Manage SKS cluster Nodepools",
	Aliases: []string{"np"},
}

func init() {
	sksCmd.AddCommand(sksNodepoolCmd)
}
