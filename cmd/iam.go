package cmd

import "github.com/spf13/cobra"

// iamCmd represent the IAM command
var iamCmd = &cobra.Command{
	Use:   "iam",
	Short: "Identity and Access Management cmd",
}

func init() {
	RootCmd.AddCommand(iamCmd)
}
