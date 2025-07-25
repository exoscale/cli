package config

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Exoscale CLI configuration management",
	RunE:  configCmdRun,
}

func configCmdRun(cmd *cobra.Command, _ []string) error {
	var (
		defaultAccountMark = promptui.Styler(promptui.FGYellow)("*")
		newAccountLabel    = "<Configure a new account>"
	)

	if exocmd.GConfigFilePath == "" && account.CurrentAccount.Key != "" {
		log.Fatalf("remove ENV credentials variables to use %s", cmd.CalledAs())
	}

	if exocmd.GConfigFilePath != "" && account.CurrentAccount.Key != "" {
		accounts := listAccounts(defaultAccountMark)
		accounts = append(accounts, newAccountLabel)
		prompt := promptui.Select{
			Label: fmt.Sprintf("Configured accounts (%s = default account)", defaultAccountMark),
			Items: accounts,
			Size:  len(accounts),
		}
		_, selectedAccount, err := prompt.Run()
		if err != nil {
			switch err {
			case promptui.ErrInterrupt:
				return nil
			default:
				return fmt.Errorf("prompt failed: %s", err)
			}
		}

		if selectedAccount == newAccountLabel {
			return addConfigAccount(false)
		}

		if strings.TrimSuffix(selectedAccount, defaultAccountMark) != account.GAllAccount.DefaultAccount {
			fmt.Printf("Setting default account to [%s]\n", selectedAccount)
			exocmd.GConfig.Set("defaultAccount", selectedAccount)
			return saveConfig(exocmd.GConfig.ConfigFileUsed(), nil)
		}

		return nil
	}

	fmt.Println("No Exoscale CLI configuration found")

	fmt.Print(`
In order to set up your configuration profile, you will need to retrieve
Exoscale API credentials from your organization's IAM:

    https://portal.exoscale.com/iam/keys

`)
	return addConfigAccount(true)
}

func saveConfig(filePath string, newAccounts *account.Config) error {
	accountsSize := 0
	currentAccounts := []account.Account{}
	if account.GAllAccount != nil {
		accountsSize = len(account.GAllAccount.Accounts)
		currentAccounts = account.GAllAccount.Accounts
	}

	newAccountsSize := 0

	if newAccounts != nil {
		newAccountsSize = len(newAccounts.Accounts)
	}

	accounts := make([]map[string]interface{}, accountsSize+newAccountsSize)

	conf := &account.Config{}

	for i, acc := range currentAccounts {
		accounts[i] = map[string]interface{}{}

		accounts[i]["name"] = acc.Name
		accounts[i]["key"] = acc.Key
		accounts[i]["defaultZone"] = acc.DefaultZone
		if acc.ClientTimeout != 0 {
			accounts[i]["clientTimeout"] = acc.ClientTimeout
		}
		if acc.DefaultOutputFormat != "" {
			accounts[i]["defaultOutputFormat"] = acc.DefaultOutputFormat
		}
		// TODO(pej): This is a workaround to not propagate Environment 'api' on config reload.
		// By default, acc.Environment is set to 'api' to be used for egoscale v2 in the whole codebase.
		// We can not tweak it like DefaultTemplate by using the 'api' default const.
		// Remove the environment when egoscale v3 will be fully integrated.
		if acc.Environment != "" && acc.Environment != "api" {
			accounts[i]["environment"] = acc.Environment
		}
		if acc.Endpoint != "" {
			accounts[i]["endpoint"] = acc.Endpoint
		}
		if acc.DefaultSSHKey != "" {
			accounts[i]["defaultSSHKey"] = acc.DefaultSSHKey
		}
		if acc.DefaultTemplate != "" {
			accounts[i]["defaultTemplate"] = acc.DefaultTemplate
		}
		if acc.Account != "" {
			accounts[i]["account"] = acc.Account
		}
		if acc.SosEndpoint != "" {
			accounts[i]["sosendpoint"] = acc.SosEndpoint
		}
		if len(acc.SecretCommand) != 0 {
			accounts[i]["secretCommand"] = acc.SecretCommand
		} else {
			accounts[i]["secret"] = acc.Secret
		}

		conf.Accounts = append(conf.Accounts, acc)
	}

	if newAccounts != nil {
		for i, acc := range newAccounts.Accounts {
			accounts[accountsSize+i] = map[string]interface{}{}

			accounts[accountsSize+i]["name"] = acc.Name
			accounts[accountsSize+i]["key"] = acc.Key
			accounts[accountsSize+i]["secret"] = acc.Secret
			accounts[accountsSize+i]["defaultZone"] = acc.DefaultZone
			conf.Accounts = append(conf.Accounts, acc)
		}
	}

	exocmd.GConfig.SetConfigType("toml")
	exocmd.GConfig.SetConfigFile(filePath)

	exocmd.GConfig.Set("accounts", accounts)

	if err := exocmd.GConfig.WriteConfig(); err != nil {
		return err
	}

	conf.DefaultAccount = exocmd.GConfig.Get("defaultAccount").(string)
	if conf.DefaultAccount == "" {
		fmt.Println("no default account set")
	}

	account.GAllAccount = conf

	return nil
}

func createConfigFile(fileName string) (string, error) {
	if _, err := os.Stat(globalstate.ConfigFolder); os.IsNotExist(err) {
		if err := os.MkdirAll(globalstate.ConfigFolder, os.ModePerm); err != nil {
			return "", err
		}
	}

	filepath := path.Join(globalstate.ConfigFolder, fileName+".toml")

	if exocmd.GConfig.ConfigFileUsed() == "" {
		if _, err := os.Stat(filepath); !os.IsNotExist(err) {
			return "", fmt.Errorf("%q exists already", filepath)
		}
	}

	fp, err := os.OpenFile(filepath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return "", err
	}
	defer fp.Close() // nolint: errcheck

	return filepath, nil
}

func listAccounts(defaultAccountMark string) []string {
	if account.GAllAccount == nil {
		return nil
	}
	res := make([]string, len(account.GAllAccount.Accounts))
	for i, acc := range account.GAllAccount.Accounts {
		res[i] = acc.Name
		if acc.Name == account.GAllAccount.DefaultAccount {
			res[i] = fmt.Sprintf("%s%s", res[i], defaultAccountMark)
		}
	}
	return res
}

func getAccountByName(name string) *account.Account {
	if account.GAllAccount == nil {
		return nil
	}

	for i, acc := range account.GAllAccount.Accounts {
		if acc.Name == name {
			return &account.GAllAccount.Accounts[i]
		}
	}

	return nil
}

func chooseZone(client *egoscale.Client, zones []string) (string, error) {
	if zones == nil {

		ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(exocmd.DefaultEnvironment, exocmd.DefaultZone))
		var err error
		zones, err = client.ListZones(ctx)

		if err != nil {
			return "", err
		}

		if len(zones) == 0 {
			return "", fmt.Errorf("no zones were found")
		}
	}

	prompt := promptui.Select{
		Label: "Default zone",
		Items: zones,
		Size:  len(zones),
	}

	_, result, err := prompt.Run()

	if err != nil {
		return "", fmt.Errorf("prompt failed %v", err)
	}

	return result, nil
}

func init() {
	exocmd.RootCmd.AddCommand(configCmd)
}
