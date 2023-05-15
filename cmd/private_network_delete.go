package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type privateNetworkDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	PrivateNetwork string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Private Network zone"`
}

func (c *privateNetworkDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *privateNetworkDeleteCmd) cmdShort() string {
	return "Delete a Private Network"
}

func (c *privateNetworkDeleteCmd) cmdLong() string { return "" }

func (c *privateNetworkDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *privateNetworkDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	privateNetwork, err := globalstate.EgoscaleClient.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetwork)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete Private Network %s?", c.PrivateNetwork)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting Private Network %s...", c.PrivateNetwork), func() {
		err = globalstate.EgoscaleClient.DeletePrivateNetwork(ctx, c.Zone, privateNetwork)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(privateNetworkCmd, &privateNetworkDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
