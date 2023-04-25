package cmd

import (
	"fmt"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type computeSSHKeyDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name string `cli-arg:"#"`

	Force bool `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *computeSSHKeyDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *computeSSHKeyDeleteCmd) cmdShort() string {
	return "Delete an SSH key"
}

func (c *computeSSHKeyDeleteCmd) cmdLong() string { return "" }

func (c *computeSSHKeyDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *computeSSHKeyDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(
		gContext,
		exoapi.NewReqEndpoint(account.CurrentAccount.Environment, account.CurrentAccount.DefaultZone),
	)

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete SSH key %s?", c.Name)) {
			return nil
		}
	}

	var err error
	decorateAsyncOperation(fmt.Sprintf("Deleting SSH key %s...", c.Name), func() {
		err = globalstate.GlobalEgoscaleClient.DeleteSSHKey(ctx, account.CurrentAccount.DefaultZone, &egoscale.SSHKey{Name: &c.Name})
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(computeSSHKeyCmd, &computeSSHKeyDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
