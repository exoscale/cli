package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:     "show <account name>",
	Short:   "Show an account details",
	Aliases: gShowAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if gAllAccount == nil {
			return fmt.Errorf("no accounts are defined")
		}

		account := gCurrentAccount.Name
		if len(args) > 0 {
			account = args[0]
		}

		if !isAccountExist(account) {
			return fmt.Errorf("account %q does not exist", args[0])
		}

		acc := getAccountByName(account)
		if acc == nil {
			return fmt.Errorf("account %q was not found", args[0])
		}

		secret := strings.Repeat("×", len(acc.Secret))

		println("Name:", acc.Name)
		println("API Key:", acc.Key)
		println("API Secret:", secret)
		println("Account:", acc.Account)
		println("Default zone:", acc.DefaultZone)
		return nil
	},
}

func init() {
	configCmd.AddCommand(showCmd)
}
