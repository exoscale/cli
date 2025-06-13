package iam

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type iamRoleListItemOutput struct {
	ID       string `json:"key"`
	Name     string `json:"name"`
	Editable bool   `json:"type"`
}

type iamRoleListOutput []iamRoleListItemOutput

func (o *iamRoleListOutput) ToJSON()  { output.JSON(o) }
func (o *iamRoleListOutput) ToText()  { output.Text(o) }
func (o *iamRoleListOutput) ToTable() { output.Table(o) }

type iamRoleListCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *iamRoleListCmd) CmdAliases() []string { return exocmd.GListAlias }

func (c *iamRoleListCmd) CmdShort() string { return "List IAM Roles" }

func (c *iamRoleListCmd) CmdLong() string {
	return fmt.Sprintf(`This command lists existing IAM Roles.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamRoleListOutput{}), ", "))
}

func (c *iamRoleListCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamRoleListCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(account.CurrentAccount.DefaultZone))
	if err != nil {
		return err
	}

	iamRoles, err := client.ListIAMRoles(ctx)
	if err != nil {
		return err
	}

	out := make(iamRoleListOutput, 0)

	for _, role := range iamRoles.IAMRoles {
		out = append(out, iamRoleListItemOutput{
			ID:       role.ID.String(),
			Name:     role.Name,
			Editable: utils.DefaultBool(role.Editable, false),
		})
	}

	return c.OutputFunc(&out, err)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(iamRoleCmd, &iamRoleListCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
