package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type privateNetworkDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	PrivateNetwork string `cli-arg:"#" cli-usage:"PRIVATE-NETWORK-NAME|ID"`

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
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	privateNetwork, err := cs.FindPrivateNetwork(ctx, c.Zone, c.PrivateNetwork)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete Private Network %s?", c.PrivateNetwork)) {
			return nil
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting Private Network %s...", c.PrivateNetwork), func() {
		err = cs.DeletePrivateNetwork(ctx, c.Zone, privateNetwork)
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
