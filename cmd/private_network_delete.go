package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type privateNetworkDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	PrivateNetwork string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"Private Network zone"`
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
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListPrivateNetworks(ctx)
	if err != nil {
		return err
	}

	pn, err := resp.FindPrivateNetwork(c.PrivateNetwork)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete Private Network %s?", c.PrivateNetwork)) {
			return nil
		}
	}

	op, err := client.DeletePrivateNetwork(ctx, pn.ID)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting Private Network %s...", c.PrivateNetwork), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
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
