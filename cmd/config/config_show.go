package config

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
)

type configShowOutput struct {
	Name               string `json:"name"`
	APIKey             string `json:"api_key"`
	APISecret          string `json:"api_secret"`
	Endpoint           string `json:"endpoint"`
	DefaultZone        string `json:"default_zone"`
	DefaultTemplate    string `json:"default_template,omitempty"`
	StorageAPIEndpoint string `json:"storage_api_endpoint,omitempty"`
	ConfigFile         string `json:"config_file" outputLabel:"Configuration File"`
	ClientTimeout      int    `json:"client_timeout" outputLabel:"API Timeout (in minutes)"`
}

func (o *configShowOutput) Type() string { return "Account" }
func (o *configShowOutput) ToJSON()      { output.JSON(o) }
func (o *configShowOutput) ToText()      { output.Text(o) }
func (o *configShowOutput) ToTable()     { output.Table(o) }

func init() {
	configCmd.AddCommand(&cobra.Command{
		Use:   "show NAME",
		Short: "Show an account details",
		Long: fmt.Sprintf(`This command shows an Exoscale account details.

Supported output template annotations: %s`,
			strings.Join(output.TemplateAnnotations(&configShowOutput{}), ", ")),
		Aliases: exocmd.GShowAlias,
		RunE: func(cmd *cobra.Command, args []string) error {
			if account.GAllAccount == nil || len(account.GAllAccount.Accounts) == 0 {
				return fmt.Errorf("no accounts configured. Run: exo config (or exo config add)")
			}

			var name string
			if len(args) > 0 {
				name = args[0]
			} else if account.CurrentAccount != nil && account.CurrentAccount.Name != "" {
				name = account.CurrentAccount.Name
			} else {
				return fmt.Errorf("default account not defined. Please specify an account name or set a default with: exo config set <account-name>")
			}

			return utils.PrintOutput(showConfig(name))
		},
	})
}

func showConfig(name string) (output.Outputter, error) {
	account := getAccountByName(name)
	if account == nil {
		return nil, fmt.Errorf("account %q was not found", name)
	}

	secret := strings.Repeat("×", len(account.Key))
	if len(account.SecretCommand) > 0 {
		secret = strings.Join(account.SecretCommand, " ")
	}

	out := configShowOutput{
		Name:               account.Name,
		ConfigFile:         exocmd.GConfigFilePath,
		Endpoint:           account.Endpoint,
		APIKey:             account.Key,
		APISecret:          secret,
		DefaultZone:        account.DefaultZone,
		DefaultTemplate:    account.DefaultTemplate,
		StorageAPIEndpoint: account.SosEndpoint,
		ClientTimeout:      account.ClientTimeout,
	}

	return &out, nil
}
