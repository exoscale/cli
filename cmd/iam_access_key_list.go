package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type iamAccessKeyListItemOutput struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type iamAccessKeyListOutput []iamAccessKeyListItemOutput

func (o *iamAccessKeyListOutput) ToJSON()  { output.JSON(o) }
func (o *iamAccessKeyListOutput) ToText()  { output.Text(o) }
func (o *iamAccessKeyListOutput) ToTable() { output.Table(o) }

type iamAccessKeyListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *iamAccessKeyListCmd) cmdAliases() []string { return gListAlias }

func (c *iamAccessKeyListCmd) cmdShort() string { return "List IAM access keys" }

func (c *iamAccessKeyListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists existing IAM access keys.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamAccessKeyListOutput{}), ", "))
}

func (c *iamAccessKeyListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamAccessKeyListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	iamAccessKeys, err := globalstate.EgoscaleClient.ListIAMAccessKeys(ctx, zone)
	if err != nil {
		return err
	}

	out := make(iamAccessKeyListOutput, 0)

	for _, k := range iamAccessKeys {
		out = append(out, iamAccessKeyListItemOutput{
			Name: *k.Name,
			Key:  *k.Key,
			Type: *k.Type,
		})
	}

	return c.outputFunc(&out, err)
}

func init() {
	cobra.CheckErr(registerCLICommand(iamAccessKeyCmd, &iamAccessKeyListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
