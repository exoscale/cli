package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type privateNetworkListItemOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type privateNetworkListOutput []privateNetworkListItemOutput

func (o *privateNetworkListOutput) toJSON()  { outputJSON(o) }
func (o *privateNetworkListOutput) toText()  { outputText(o) }
func (o *privateNetworkListOutput) toTable() { outputTable(o) }

type privateNetworkListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *privateNetworkListCmd) cmdAliases() []string { return gListAlias }

func (c *privateNetworkListCmd) cmdShort() string { return "List Private Networks" }

func (c *privateNetworkListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Compute instance Private Networks.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&privateNetworkListItemOutput{}), ", "))
}

func (c *privateNetworkListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *privateNetworkListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	privateNetworks, err := cs.ListPrivateNetworks(ctx, gCurrentAccount.DefaultZone)
	if err != nil {
		return err
	}

	out := make(privateNetworkListOutput, 0)

	for _, t := range privateNetworks {
		out = append(out, privateNetworkListItemOutput{
			ID:   *t.ID,
			Name: *t.Name,
		})
	}

	return output(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(privateNetworkCmd, &privateNetworkListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
