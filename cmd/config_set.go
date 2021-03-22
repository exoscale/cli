package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configSetCmd = &cobra.Command{
	Use:   "set NAME",
	Short: "Set an account as default account",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return cmd.Usage()
		}
		if gAllAccount == nil {
			return fmt.Errorf("no accounts configured")
		}

		if a := getAccountByName(args[0]); a == nil {
			return fmt.Errorf("account %q does not exist", args[0])
		}

		gConfig.Set("defaultAccount", args[0])

		if err := saveConfig(gConfig.ConfigFileUsed(), nil); err != nil {
			return err
		}

		fmt.Printf("Default profile set to [%s]\n", args[0])

		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
}
