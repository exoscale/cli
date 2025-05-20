package load_balancer

import (
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type nlbDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *nlbDeleteCmd) CmdShort() string { return "Delete a Network Load Balancer" }

func (c *nlbDeleteCmd) CmdLong() string { return "" }

func (c *nlbDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	if !c.Force {
		if !utils.AskQuestion(
			ctx,
			fmt.Sprintf(
				"Are you sure you want to delete Network Load Balancer %q?",
				c.NetworkLoadBalancer,
			)) {
			return nil
		}
	}

	nlb, err := globalstate.EgoscaleClient.FindNetworkLoadBalancer(ctx, c.Zone, c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting Network Load Balancer %q...", c.NetworkLoadBalancer), func() {
		err = globalstate.EgoscaleClient.DeleteNetworkLoadBalancer(ctx, c.Zone, nlb)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(nlbCmd, &nlbDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
