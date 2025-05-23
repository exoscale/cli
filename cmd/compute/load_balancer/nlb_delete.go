package load_balancer

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type nlbDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *nlbDeleteCmd) CmdShort() string { return "Delete a Network Load Balancer" }

func (c *nlbDeleteCmd) CmdLong() string { return "" }

func (c *nlbDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {

	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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
		if !utils.AskQuestion(ctx,
			fmt.Sprintf(
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
	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting Network Load Balancer %q...", c.NetworkLoadBalancer), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	return err
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(nlbCmd, &nlbDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
