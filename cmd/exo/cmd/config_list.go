package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available accounts",
	Run: func(cmd *cobra.Command, args []string) {
		if allAccount == nil {
			log.Fatalf("No accounts defined")
		}
		listAccounts()
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
}
