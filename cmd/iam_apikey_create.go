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

	newAccount := &account{
		Endpoint: defaultEndpoint,
		Key:      apiKey.Key,
		Secret:   apiKey.Secret,
		Name:     apiKey.Name,
	}

	if askQuestion("do you wish to add this account in your config file?") {
		resp, err := cs.GetWithContext(gContext, egoscale.Account{})
		if err != nil {
			return err
		}

		acc := resp.(*egoscale.Account)
		newAccount.Account = acc.Name

		endpoint, err := readInput(reader, "API Endpoint", newAccount.Endpoint)
		if err != nil {
			return err
		}
		if endpoint != newAccount.Endpoint {
			newAccount.Endpoint = endpoint
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

		zonesResp, err := cs.ListWithContext(gContext, &egoscale.Zone{ID: acc.DefaultZoneID})
		if err != nil {
			return err
		}

		zone := zonesResp[0].(*egoscale.Zone)
		zName := strings.ToLower(zone.Name)

		newAccount.DefaultZone = zName
		newAccount.DNSEndpoint = strings.Replace(newAccount.Endpoint, "/compute", "/dns", 1)

		config.Accounts = append(config.Accounts, *newAccount)
		if askQuestion("Make [" + newAccount.Name + "] your default profile?") {
			config.DefaultAccount = newAccount.Name
			viper.Set("defaultAccount", newAccount.Name)
		}

		return addAccount(viper.ConfigFileUsed(), config)
	}
	return nil
}

func init() {
	apiKeyCreateCmd.Flags().StringSliceP("operation", "o", []string{}, "API key allowed operation")
	apiKeyCmd.AddCommand(apiKeyCreateCmd)
}
