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
		if len(args) < 1 {
			return cmd.Usage()
		}
		if gAllAccount == nil {
			return fmt.Errorf("no accounts are defined")
		}
		if !isAccountExist(args[0]) {
			return fmt.Errorf("account %q does not exist", args[0])
		}
		acc := getAccountByName(args[0])
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
