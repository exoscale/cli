package config

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
)

var configSetCmd = &cobra.Command{
	Use:   "set NAME",
	Short: "Set an account as default account",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		if account.GAllAccount == nil || len(account.GAllAccount.Accounts) == 0 {
			return fmt.Errorf("no accounts configured. Run: exo config")
		}

		if a := getAccountByName(args[0]); a == nil {
			return fmt.Errorf("account %q does not exist", args[0])
		}

		exocmd.GConfig.Set("defaultAccount", args[0])

		if err := saveConfig(exocmd.GConfig.ConfigFileUsed(), nil); err != nil {
			return err
		}

		fmt.Printf("Default profile set to [%s]\n", args[0])

		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
}
