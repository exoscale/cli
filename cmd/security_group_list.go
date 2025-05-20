package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/exoscale/egoscale/v2/oapi"
)

type securityGroupListItemOutput struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
}

type securityGroupListOutput []securityGroupListItemOutput

func (o *securityGroupListOutput) ToJSON()  { output.JSON(o) }
func (o *securityGroupListOutput) ToText()  { output.Text(o) }
func (o *securityGroupListOutput) ToTable() { output.Table(o) }

type securityGroupListCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Visibility string `cli-usage:"Security Group visibility: private (default) or public"`
}

func (c *securityGroupListCmd) CmdAliases() []string { return GListAlias }

func (c *securityGroupListCmd) CmdShort() string { return "List Security Groups" }

func (c *securityGroupListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Compute instance Security Groups.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&securityGroupListItemOutput{}), ", "))
}

func (c *securityGroupListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		GContext,
		exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone),
	)

	params := &oapi.ListSecurityGroupsParams{}
	if len(c.Visibility) > 0 {
		params = &oapi.ListSecurityGroupsParams{
			Visibility: (*oapi.ListSecurityGroupsParamsVisibility)(&c.Visibility),
		}
	}
	securityGroups, err := globalstate.EgoscaleClient.FindSecurityGroups(ctx, account.CurrentAccount.DefaultZone, params)
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

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(RegisterCLICommand(securityGroupCmd, &securityGroupListCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
