package cmd

import "github.com/spf13/cobra"

// apiKeycmd represent the API key command
var apiKeyCmd = &cobra.Command{
	Use:   "apikey",
	Short: "API Keys management",
}

func init() {
	iamCmd.AddCommand(apiKeyCmd)
}
