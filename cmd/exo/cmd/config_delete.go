package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deleteCmd represents the delete command
var configDeleteCmd = &cobra.Command{
	Use:     "delete <account name>",
	Short:   "Delete an account from config file",
	Aliases: gDeleteAlias,
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

		if args[0] == gAllAccount.DefaultAccount {
			log.Fatalf("Can't delete a default account")
		}

		pos := 0
		for i, acc := range gAllAccount.Accounts {
			if acc.Name == args[0] {
				pos = i
				break
			}
		}

		gAllAccount.Accounts = append(gAllAccount.Accounts[:pos], gAllAccount.Accounts[pos+1:]...)

		if err := addAccount(viper.ConfigFileUsed(), nil); err != nil {
			log.Fatal(err)
		}

		println(args[0])
	},
}

func init() {
	configCmd.AddCommand(configDeleteCmd)
}
