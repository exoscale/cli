package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type securityGroupListItemOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type securityGroupListOutput []securityGroupListItemOutput

func (o *securityGroupListOutput) toJSON()  { outputJSON(o) }
func (o *securityGroupListOutput) toText()  { outputText(o) }
func (o *securityGroupListOutput) toTable() { outputTable(o) }

type securityGroupListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *securityGroupListCmd) cmdAliases() []string { return gListAlias }

func (c *securityGroupListCmd) cmdShort() string { return "List Security Groups" }

func (c *securityGroupListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists Compute instance Security Groups.

Supported output template annotations: %s`,
		strings.Join(outputterTemplateAnnotations(&securityGroupListItemOutput{}), ", "))
}

func (c *securityGroupListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(gCurrentAccount.Environment, gCurrentAccount.DefaultZone),
	)

	securityGroups, err := cs.ListSecurityGroups(ctx, gCurrentAccount.DefaultZone)
	if err != nil {
		return err
	}

	out := make(securityGroupListOutput, 0)

	for _, t := range securityGroups {
		out = append(out, securityGroupListItemOutput{
			ID:   *t.ID,
			Name: *t.Name,
		})
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(securityGroupCmd, &securityGroupListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
