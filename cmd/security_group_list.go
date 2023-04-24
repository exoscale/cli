package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
	"github.com/spf13/cobra"
)

type securityGroupListItemOutput struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
}

type securityGroupListOutput []securityGroupListItemOutput

func (o *securityGroupListOutput) toJSON()  { output.JSON(o) }
func (o *securityGroupListOutput) toText()  { output.Text(o) }
func (o *securityGroupListOutput) toTable() { output.Table(o) }

type securityGroupListCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Visibility string `cli-usage:"Security Group visibility: private (default) or public"`
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

	params := &oapi.ListSecurityGroupsParams{}
	if len(c.Visibility) > 0 {
		params = &oapi.ListSecurityGroupsParams{
			Visibility: (*oapi.ListSecurityGroupsParamsVisibility)(&c.Visibility),
		}
	}
	securityGroups, err := cs.FindSecurityGroups(ctx, gCurrentAccount.DefaultZone, params)
	if err != nil {
		return err
	}

	out := make(securityGroupListOutput, 0)

	for _, t := range securityGroups {
		sg := securityGroupListItemOutput{Name: *t.Name}
		if t.ID != nil {
			sg.ID = *t.ID
			sg.Visibility = "private"
		} else {
			sg.Visibility = "public"
		}
		out = append(out, sg)
	}

	return c.outputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(registerCLICommand(securityGroupCmd, &securityGroupListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
