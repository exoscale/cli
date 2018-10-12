package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// setCmd represents the set command
var configSetCmd = &cobra.Command{
	Use:   "set [account name]",
	Short: "Set an account as default",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return cmd.Usage()
		}

		if gAllAccount == nil {
			return fmt.Errorf("no accounts are defined")
		}

		var result string

		if len(args) == 1 {
			result = args[0]
		}
		if len(args) == 0 {

			accounts := make([]string, len(gAllAccount.Accounts))

			for i, account := range gAllAccount.Accounts {
				accounts[i] = account.Name
				if account.Name == gAllAccount.DefaultAccount {
					accounts[i] = fmt.Sprintf("%s [Default]", account.Name)
				}
			}

			prompt := promptui.Select{
				Label: "Select an account",
				Items: accounts,
			}

			var err error
			_, result, err = prompt.Run()

			if err != nil {
				return fmt.Errorf("Prompt failed %v", err)
			}

			if fmt.Sprintf("%s [Default]", gAllAccount.DefaultAccount) == result {
				return nil
			}

		}

		if !isAccountExist(result) {
			return fmt.Errorf("account %q does not exist", args[0])
		}

		viper.Set("defaultAccount", result)

		if err := addAccount(viper.ConfigFileUsed(), nil); err != nil {
			return err
		}

		println("Default profile set to", result)

		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
}
