package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type nlbDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *nlbDeleteCmd) cmdShort() string { return "Delete a Network Load Balancer" }

func (c *nlbDeleteCmd) cmdLong() string { return "" }

func (c *nlbDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	if !c.Force {
		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to delete Network Load Balancer %q?",
			c.NetworkLoadBalancer,
		)) {
			return nil
		}
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	nlb, err := globalstate.EgoscaleClient.FindNetworkLoadBalancer(ctx, c.Zone, c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting Network Load Balancer %q...", c.NetworkLoadBalancer), func() {
		err = globalstate.EgoscaleClient.DeleteNetworkLoadBalancer(ctx, c.Zone, nlb)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbCmd, &nlbDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
