package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type iamAccessKeyListItemOutput struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type iamAccessKeyListOutput []iamAccessKeyListItemOutput

func (o *iamAccessKeyListOutput) toJSON()  { outputJSON(o) }
func (o *iamAccessKeyListOutput) toText()  { outputText(o) }
func (o *iamAccessKeyListOutput) toTable() { outputTable(o) }

type iamAccessKeyListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *iamAccessKeyListCmd) cmdAliases() []string { return gListAlias }

func (c *iamAccessKeyListCmd) cmdShort() string { return "List IAM access keys" }

func (c *iamAccessKeyListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists existing IAM access keys.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&iamAccessKeyListOutput{}), ", "))
}

func (c *iamAccessKeyListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamAccessKeyListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := gCurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, zone))

	iamAccessKeys, err := cs.ListIAMAccessKeys(ctx, zone)
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
