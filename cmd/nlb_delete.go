package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type nlbDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *nlbDeleteCmd) cmdShort() string { return "Delete a Network Load Balancer" }

func (c *nlbDeleteCmd) cmdLong() string { return "" }

func (c *nlbDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {

	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	nlbs, err := client.ListLoadBalancers(ctx)
	if err != nil {
		return err
	}

	nlb, err := nlbs.FindLoadBalancer(c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf(
			"Are you sure you want to delete Network Load Balancer %q?",
			nlb.ID,
		)) {
			return nil
		}
	}

	op, err := client.DeleteLoadBalancer(ctx, nlb.ID)
	if err != nil {
		return err
	}
	decorateAsyncOperation(fmt.Sprintf("Deleting Network Load Balancer %q...", c.NetworkLoadBalancer), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	return err
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbCmd, &nlbDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
