package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/table"
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

		name := gCurrentAccount.AccountName()
		if len(args) > 0 {
			name = args[0]
		}

		if !doesAccountExist(name) {
			return fmt.Errorf("account %q does not exist", name)
		}

		account := getAccountByName(name)
		if account == nil {
			return fmt.Errorf("account %q was not found", name)
		}

		secret := strings.Repeat("×", len(account.Key))
		if len(account.SecretCommand) > 0 {
			secret = strings.Join(account.SecretCommand, " ")
		}

		t := table.NewTable(os.Stdout)
		t.SetHeader([]string{name})

		t.Append([]string{"Account Name", account.Account})
		t.Append([]string{"API Key", account.Key})
		t.Append([]string{"API Secret", secret})
		t.Append([]string{"Default Zone", account.DefaultZone})

		if account.DefaultTemplate != "" {
			t.Append([]string{"Default Template", account.DefaultTemplate})
		}

		if account.Endpoint != defaultEndpoint {
			t.Append([]string{"Endpoint", account.Endpoint})
			t.Append([]string{"DNS Endpoint", account.DNSEndpoint})
		}

		if account.SosEndpoint != defaultSosEndpoint {
			t.Append([]string{"SOS Endpoint", account.SosEndpoint})
		}

		if gConfigFilePath != "" {
			t.Append([]string{"Configuration Folder", gConfigFolder})
			t.Append([]string{"Configuration", gConfigFilePath})
		} else {
			t.Append([]string{"Configuration", "environment variables"})
		}

		t.Render()

		return nil
	},
}

func init() {
	configCmd.AddCommand(showCmd)
}
