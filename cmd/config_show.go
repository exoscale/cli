package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:     "show <account name>",
	Short:   "Show an account details",
	Aliases: gShowAlias,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Usage()
			return
		}
		if gAllAccount == nil {
			log.Fatalf("No accounts defined")
		}
		if !isAccountExist(args[0]) {
			log.Fatalf("Account %q doesn't exist", args[0])
		}
		acc := getAccountByName(args[0])
		if acc == nil {
			log.Fatalf("Account %q not found", args[0])
		}

		println("Name:", acc.Name)
		println("API Key:", acc.Key)
		println("API Secret:", acc.Secret)
		println("Account:", acc.Account)
		println("Default zone:", acc.DefaultZone)

	},
}

func init() {
	configCmd.AddCommand(showCmd)
}
