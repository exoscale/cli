package cmd

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type iamRoleDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Role string `cli-arg:"#" cli-usage:"ID|NAME"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *iamRoleDeleteCmd) cmdAliases() []string { return gDeleteAlias }

func (c *iamRoleDeleteCmd) cmdShort() string {
	return "Delete IAM Role"
}

func (c *iamRoleDeleteCmd) cmdLong() string {
	return `This commands delete existing API Role.
Command will fail if Role is attached to an API Key.`
}

func (c *iamRoleDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *iamRoleDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	zone := account.CurrentAccount.DefaultZone
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, zone))

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete IAM Role %s?", c.Role)) {
			return nil
		}
	}

	var err error
	decorateAsyncOperation(fmt.Sprintf("Deleting IAM role %s...", c.Role), func() {
		if _, err = uuid.Parse(c.Role); err != nil {
			roles, err := globalstate.EgoscaleClient.ListIAMRoles(ctx, zone)
			if err != nil {
				return
			}

			for _, role := range roles {
				if role.Name != nil && *role.Name == c.Role {
					c.Role = *role.ID
					break
				}
			}
		}

		err = globalstate.EgoscaleClient.DeleteIAMRole(ctx, zone, &egoscale.IAMRole{ID: &c.Role})
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(iamRoleCmd, &iamRoleDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
