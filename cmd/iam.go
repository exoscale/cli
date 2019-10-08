package cmd

import "github.com/spf13/cobra"

var iamCmd = &cobra.Command{
	Use:   "iam",
	Short: "Identity and Access Management cmd",
}

func init() {
	RootCmd.AddCommand(iamCmd)
}
