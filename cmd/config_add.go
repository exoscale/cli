package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
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

			config := &config{Accounts: []account{*newAccount}}
			if askQuestion("Set [" + newAccount.Name + "] as default account?") {
				config.DefaultAccount = newAccount.Name
				gConfig.Set("defaultAccount", newAccount.Name)
			}

			return saveConfig(gConfig.ConfigFileUsed(), config)
		},
	})
}

func addConfigAccount(firstRun bool) error {
	var (
		config config
		err    error
	)

	filePath := gConfig.ConfigFileUsed()

	if firstRun {
		if filePath, err = createConfigFile(defaultConfigFileName); err != nil {
			return err
		}

		gConfig.SetConfigFile(filePath)
	}

	newAccount, err := promptAccountInformation()
	if err != nil {
		return err
	}
	config.DefaultAccount = newAccount.Name
	config.Accounts = []account{*newAccount}
	gConfig.Set("defaultAccount", newAccount.Name)

	if len(config.Accounts) == 0 {
		return nil
	}

	return saveConfig(filePath, &config)
}

func promptAccountInformation() (*account, error) {
	var client *egoscale.Client

	reader := bufio.NewReader(os.Stdin)
	account := &account{
		Endpoint: defaultEndpoint,
		Key:      "",
		Secret:   "",
	}

	for i := 0; ; i++ {
		if i > 0 {
			endpoint, err := readInput(reader, "API Endpoint", account.Endpoint)
			if err != nil {
				return nil, err
			}
			if endpoint != account.Endpoint {
				account.Endpoint = endpoint
			}
		}

		apiKey, err := readInput(reader, "API Key", account.Key)
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
		secretKey, err := readInput(reader, "Secret Key", secretShow)
		if err != nil {
			return nil, err
		}
		if secretKey != secret && secretKey != secretShow {
			account.Secret = secretKey
		}

		client = egoscale.NewClient(account.Endpoint, account.Key, account.APISecret())

		fmt.Printf("Retrieving account information...")
		resp, err := client.GetWithContext(gContext, egoscale.Account{})
		if err != nil {
			if egoerr, ok := err.(*egoscale.ErrorResponse); ok && egoerr.ErrorCode == egoscale.ErrorCode(403) {
				fmt.Print(`

Unable to retrieve information, please enter your account details:

`)
				for {
					acc, err := readInput(reader, "Account name", account.Account)
					if err != nil {
						return nil, err
					}
					if acc != "" {
						account.Account = acc
						break
					}
				}

				break
			}

			fmt.Print(` failure.

Let's start over.

`)
		} else {
			fmt.Print(" done!\n\n")
			acc := resp.(*egoscale.Account)
			account.Name = acc.Name
			account.Account = acc.Name
			break
		}
	}

	name, err := readInput(reader, "Name", account.Name)
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
		name, err = readInput(reader, "Name", account.Name)
		if err != nil {
			return nil, err
		}

		account.Name = name
	}

	account.DefaultZone, err = chooseZone(client, nil)
	if err != nil {
		if egoerr, ok := err.(*egoscale.ErrorResponse); ok && egoerr.ErrorCode == egoscale.ErrorCode(403) {
			for {
				defaultZone, err := chooseZone(cs, zones)
				if err != nil {
					return nil, err
				}
				if defaultZone != "" {
					account.DefaultZone = defaultZone
					break
				}
			}
		} else {
			return nil, err
		}
	}

	account.DNSEndpoint = buildDNSAPIEndpoint(account.Endpoint)

	return account, nil
}
