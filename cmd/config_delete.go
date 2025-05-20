package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
)

var configDeleteCmd = &cobra.Command{
	Use:     "delete NAME",
	Short:   "Delete an account from configuration",
	Aliases: GDeleteAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		if account.GAllAccount == nil {
			return fmt.Errorf("no accounts defined")
		}
		if a := getAccountByName(args[0]); a == nil {
			return fmt.Errorf("account %q doesn't exist", args[0])
		}

		if args[0] == account.GAllAccount.DefaultAccount {
			return fmt.Errorf("cannot delete the default account")
		}

		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}

		if !force {
			if !askQuestion(fmt.Sprintf("Are you sure you want to delete the account %q from configuration?", args[0])) {
				return nil
			}
		}

		pos := 0
		for i, acc := range account.GAllAccount.Accounts {
			if acc.Name == args[0] {
				pos = i
				break
			}
		}

		account.GAllAccount.Accounts = append(account.GAllAccount.Accounts[:pos], account.GAllAccount.Accounts[pos+1:]...)

		if err := saveConfig(gConfig.ConfigFileUsed(), nil); err != nil {
			return err
		}

		println(args[0])
		return nil
	},
}

func init() {
	configDeleteCmd.Flags().BoolP("force", "f", false, cmdFlagForceHelp)
	configCmd.AddCommand(configDeleteCmd)
}
