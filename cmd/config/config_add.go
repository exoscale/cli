package config

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	exo "github.com/exoscale/egoscale/v2"
)

func init() {
	configCmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add a new account to configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			newAccount, err := promptAccountInformation()
			if err != nil {
				return err
			}

			config := &account.Config{Accounts: []account.Account{*newAccount}}
			if utils.AskQuestion(exocmd.GContext, "Set ["+newAccount.Name+"] as default account?") {
				config.DefaultAccount = newAccount.Name
				exocmd.GConfig.Set("defaultAccount", newAccount.Name)
			}

			return saveConfig(exocmd.GConfig.ConfigFileUsed(), config)
		},
	})
}

func addConfigAccount(firstRun bool) error {
	var (
		config account.Config
		err    error
	)

	filePath := exocmd.GConfig.ConfigFileUsed()

	if firstRun {
		if filePath, err = createConfigFile(exocmd.DefaultConfigFileName); err != nil {
			return err
		}

		exocmd.GConfig.SetConfigFile(filePath)
	}

	newAccount, err := promptAccountInformation()
	if err != nil {
		return err
	}
	config.DefaultAccount = newAccount.Name
	config.Accounts = []account.Account{*newAccount}
	exocmd.GConfig.Set("defaultAccount", newAccount.Name)

	if len(config.Accounts) == 0 {
		return nil
	}

	return saveConfig(filePath, &config)
}

func promptAccountInformation() (*account.Account, error) {
	var client *exo.Client

	ctx := exocmd.GContext

	reader := bufio.NewReader(os.Stdin)
	account := &account.Account{}

	apiKey, err := utils.ReadInput(ctx, reader, "API Key", account.Key)
	if err != nil {
		return nil, err
	}
	if apiKey != account.Key {
		account.Key = apiKey
	}

	secret := account.APISecret()
	secretShow := account.APISecret()
	if secret != "" && len(secret) > 10 {
		secretShow = secret[0:7] + "..."
	}
	secretKey, err := utils.ReadInput(ctx, reader, "Secret Key", secretShow)
	if err != nil {
		return nil, err
	}
	if secretKey != secret && secretKey != secretShow {
		account.Secret = secretKey
	}

	name, err := utils.ReadInput(ctx, reader, "Name", account.Name)
	if err != nil {
		return nil, err
	}
	if name != "" {
		account.Name = name
	}

	for {
		if a := getAccountByName(account.Name); a == nil {
			break
		}

		fmt.Printf("Name [%s] already exist\n", name)
		name, err = utils.ReadInput(ctx, reader, "Name", account.Name)
		if err != nil {
			return nil, err
		}

		account.Name = name
	}

	client, err = exo.NewClient(account.Key, account.APISecret())
	if err != nil {
		return nil, err
	}
	account.DefaultZone, err = chooseZone(client, nil)
	if err != nil {
		for {
			defaultZone, err := chooseZone(globalstate.EgoscaleClient, utils.AllZones)
			if err != nil {
				return nil, err
			}
			if defaultZone != "" {
				account.DefaultZone = defaultZone
				break
			}
		}
	}

	return account, nil
}
