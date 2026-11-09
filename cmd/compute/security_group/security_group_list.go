package security_group

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/clicmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
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
	clicmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`

	Visibility v3.ListSecurityGroupsVisibility `cli-usage:"Security Group visibility: private (default) or public"`
}

func (c *securityGroupListCmd) CmdAliases() []string { return clicmd.GListAlias }

func (c *securityGroupListCmd) CmdShort() string { return "List Security Groups" }

func (c *securityGroupListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists Compute instance Security Groups.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&securityGroupListItemOutput{}), ", "))
}

func (c *securityGroupListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return clicmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *securityGroupListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	securityGroups, err := func() (*v3.ListSecurityGroupsResponse, error) {
		if c.Visibility == "" {
			return client.ListSecurityGroups(ctx)

		} else {
			return client.ListSecurityGroups(ctx, v3.ListSecurityGroupsWithVisibility(c.Visibility))
		}
	}()
	if err != nil {
		return err
	}

	out := make(securityGroupListOutput, 0)

	for _, t := range securityGroups.SecurityGroups {
		sg := securityGroupListItemOutput{Name: t.Name}
		if t.ID != "" {
			sg.ID = t.ID.String()
			sg.Visibility = "private"
		} else {
			sg.Visibility = "public"
		}
		out = append(out, sg)
	}

	return c.OutputFunc(&out, nil)
}

func init() {
	cobra.CheckErr(clicmd.RegisterCLICommand(securityGroupCmd, &securityGroupListCmd{
		CliCommandSettings: clicmd.DefaultCLICmdSettings(),
	}))
}
