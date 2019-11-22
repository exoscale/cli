package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/egoscale"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	configCmd.AddCommand(&cobra.Command{
		Use:   "add <account name>",
		Short: "Add a new account to configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return cmd.Usage()
			}

			newAccount, err := promptAccountInformation()
			if err != nil {
				return err
			}

			config := &config{Accounts: []account{*newAccount}}
			if askQuestion("Set [" + newAccount.Name + "] as default account?") {
				config.DefaultAccount = newAccount.Name
				viper.Set("defaultAccount", newAccount.Name)
			}

			return saveConfig(viper.ConfigFileUsed(), config)
		},
	})
}

func addConfigAccount(firstRun bool) error {
	var config config

	if firstRun {
		filePath, err := createConfigFile(defaultConfigFileName)
		if err != nil {
			return err
		}

		viper.SetConfigFile(filePath)

		newAccount, err := promptAccountInformation()
		if err != nil {
			return err
		}
		config.DefaultAccount = newAccount.Name
		config.Accounts = []account{*newAccount}
		viper.Set("defaultAccount", newAccount.Name)
	}

	if len(config.Accounts) == 0 {
		return nil
	}

	return saveConfig(viper.ConfigFileUsed(), &config)
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
				fmt.Print(`failure.

Please enter your account information.

`)
				for {
					acc, err := readInput(reader, "Account", account.Account)
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

	account.DefaultZone, err = chooseZone(account.Name, client)
	if err != nil {
		if egoerr, ok := err.(*egoscale.ErrorResponse); ok && egoerr.ErrorCode == egoscale.ErrorCode(403) {
			for {
				defaultZone, err := readInput(reader, "Zone", account.DefaultZone)
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

	account.DNSEndpoint = strings.Replace(account.Endpoint, "/compute", "/dns", 1)

	return account, nil
}
