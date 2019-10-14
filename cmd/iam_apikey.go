package cmd

import "github.com/spf13/cobra"

// addCmd represents the add command
var iamAPIKeyCmd = &cobra.Command{
	Use:   "apikey",
	Short: "API Keys management",
}

func init() {
	iamCmd.AddCommand(iamAPIKeyCmd)
}
