package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/table"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type iamAccessKeyShowOutput struct {
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	APIKey     string    `json:"api_key"`
	APISecret  *string   `json:"api_secret,omitempty"`
	Operations *[]string `json:"operations,omitempty"`
	Tags       *[]string `json:"tags,omitempty"`
	Resources  *[]string `json:"resources,omitempty"`
}

func (o *iamAccessKeyShowOutput) ToJSON() { output.JSON(o) }
func (o *iamAccessKeyShowOutput) ToText() { output.Text(o) }
func (o *iamAccessKeyShowOutput) ToTable() {
	if o.APISecret != nil {
		defer fmt.Fprint(os.Stderr, `
/!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\
/!\    Ensure to save your API Secret somewhere,    /!\
/!\   as there is no way to recover it afterwards   /!\
/!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\ /!\
`)
	}

	t := table.NewTable(os.Stdout)
	t.SetHeader([]string{"IAM Access Key"})
	defer t.Render()

	t.Append([]string{"Name", o.Name})
	t.Append([]string{"Type", o.Type})
	t.Append([]string{"API Key", o.APIKey})
	t.Append([]string{"API Secret", utils.DefaultString(o.APISecret, strings.Repeat("*", 43))})

	if o.Operations != nil {
		t.Append([]string{"Operations", strings.Join(*o.Operations, "\n")})
	}

	if o.Tags != nil {
		t.Append([]string{"Tags", strings.Join(*o.Tags, "\n")})
	}

	if o.Resources != nil {
		t.Append([]string{"Resources", strings.Join(*o.Resources, "\n")})
	}
}

type iamAccessKeyShowCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	APIKey string `cli-arg:"#" cli-usage:"API-KEY"`
}

func (c *iamAccessKeyShowCmd) cmdAliases() []string { return gShowAlias }

func (c *iamAccessKeyShowCmd) cmdShort() string {
	return "Show an IAM access key details"
}

func (c *iamAccessKeyShowCmd) cmdLong() string {
	return fmt.Sprintf(`This command shows an IAM access key details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamAccessKeyShowOutput{}), ", "))
}

func (c *iamAccessKeyShowCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamAccessKeyShowCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	iamAccessKey, err := globalstate.EgoscaleClient.GetIAMAccessKey(ctx, zone, c.APIKey)
	if err != nil {
		return err
	}

	out := iamAccessKeyShowOutput{
		Name:       *iamAccessKey.Name,
		APIKey:     *iamAccessKey.Key,
		Operations: iamAccessKey.Operations,
		Resources: func() *[]string {
			if iamAccessKey.Resources != nil {
				list := make([]string, len(*iamAccessKey.Resources))
				for i, r := range *iamAccessKey.Resources {
					list[i] = fmt.Sprintf("%s/%s:%s", r.Domain, r.ResourceType, r.ResourceName)
				}
				return &list
			}
			return nil
		}(),
		Tags: iamAccessKey.Tags,
		Type: *iamAccessKey.Type,
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(iamAccessKeyCmd, &iamAccessKeyShowCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
