package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
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
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"list"`
}

func (c *iamRoleListCmd) cmdAliases() []string { return gListAlias }

func (c *iamRoleListCmd) cmdShort() string { return "List IAM Roles" }

func (c *iamRoleListCmd) cmdLong() string {
	return fmt.Sprintf(`This command lists existing IAM Roles.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&iamRoleListOutput{}), ", "))
}

func (c *iamRoleListCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamRoleListCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	iamRoles, err := globalstate.EgoscaleClient.ListIAMRoles(ctx, zone)
	if err != nil {
		return err
	}

	out := make(iamRoleListOutput, 0)

	for _, role := range iamRoles {
		out = append(out, iamRoleListItemOutput{
			ID:       utils.DefaultString(role.ID, ""),
			Name:     utils.DefaultString(role.Name, ""),
			Editable: utils.DefaultBool(role.Editable, false),
		})
	}

	return c.outputFunc(&out, err)
}

func init() {
	cobra.CheckErr(registerCLICommand(iamRoleCmd, &iamRoleListCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
