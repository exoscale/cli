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

type apiKeyCreateItemOutput struct {
	Name       string   `json:"name"`
	Key        string   `json:"key"`
	Secret     string   `json:"secret,omitempty"`
	Operations []string `json:"operations,omitempty"`
	Type       string   `json:"type"`
}

func (o *apiKeyCreateItemOutput) toJSON()  { outputJSON(o) }
func (o *apiKeyCreateItemOutput) toText()  { outputText(o) }
func (o *apiKeyCreateItemOutput) toTable() { outputTable(o) }

// apiKeyCreateCmd represents an API key creation command
var apiKeyCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create an API key",
	Long: fmt.Sprintf(`This command create an API key.

	Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&apiKeyCreateItemOutput{}), ", ")),
	Aliases: gCreateAlias,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return cmd.Usage()
		}

		ops, err := cmd.Flags().GetStringSlice("operation")
		if err != nil {
			return err
		}

		resp, err := cs.RequestWithContext(gContext, &egoscale.CreateAPIKey{
			Name:       args[0],
			Operations: strings.Join(ops, ","),
		})
		if err != nil {
			return err
		}

		apiKey := resp.(*egoscale.APIKey)

		if !gQuiet {
			o := apiKeyCreateItemOutput{
				Name:       apiKey.Name,
				Key:        apiKey.Key,
				Secret:     apiKey.Secret,
				Operations: apiKey.Operations,
				Type:       string(apiKey.Type),
			}

			if err := output(&o, err); err != nil {
				return err
			}
		}

		return addAPIKeyInConfigFile(apiKey)
	},
}

func addAPIKeyInConfigFile(apiKey *egoscale.APIKey) error {
	reader := bufio.NewReader(os.Stdin)

	config := &config{}

	var client *egoscale.Client

	newAccount := &account{
		Endpoint: defaultEndpoint,
		Key:      apiKey.Key,
		Secret:   apiKey.Secret,
	}

	if askQuestion("do you wish to add this account in your config file?") {
		for i := 0; ; i++ {
			if i > 0 {
				endpoint, err := readInput(reader, "API Endpoint", newAccount.Endpoint)
				if err != nil {
					return err
				}
				if endpoint != newAccount.Endpoint {
					newAccount.Endpoint = endpoint
				}
			}

			client = egoscale.NewClient(newAccount.Endpoint, newAccount.Key, newAccount.APISecret())

			fmt.Printf("Checking the credentials of %q...", newAccount.Key)
			resp, err := client.GetWithContext(gContext, egoscale.Account{})
			if err != nil {
				fmt.Print(` failure.

Let's start over.

`)
			} else {
				fmt.Print(" success!\n\n")
				acc := resp.(*egoscale.Account)
				newAccount.Name = apiKey.Name
				newAccount.Account = acc.Name
				break
			}
		}
	}

	name, err := readInput(reader, "Name", newAccount.Name)
	if err != nil {
		return err
	}
	if name != "" {
		newAccount.Name = name
	}

	for {
		if a := getAccountByName(newAccount.Name); a == nil {
			break
		}

		fmt.Printf("Name [%s] already exist\n", name)
		name, err = readInput(reader, "Name", newAccount.Name)
		if err != nil {
			return err
		}

		newAccount.Name = name
	}

	defaultZone, err := chooseZone(newAccount.Name, client)
	if err != nil {
		return err
	}

	newAccount.DefaultZone = defaultZone
	newAccount.DNSEndpoint = strings.Replace(newAccount.Endpoint, "/compute", "/dns", 1)

	config.Accounts = append(config.Accounts, *newAccount)
	if askQuestion("Make [" + newAccount.Name + "] your default profile?") {
		config.DefaultAccount = newAccount.Name
		viper.Set("defaultAccount", newAccount.Name)
	}

	return addAccount(viper.ConfigFileUsed(), config)
}

func init() {
	apiKeyCreateCmd.Flags().StringSliceP("operation", "o", []string{}, "API key allowed operation")
	apiKeyCmd.AddCommand(apiKeyCreateCmd)
}
