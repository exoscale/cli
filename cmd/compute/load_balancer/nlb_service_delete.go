package load_balancer

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type nlbServiceDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"LOAD-BALANCER-NAME|ID"`
	Service             string `cli-arg:"#" cli-usage:"SERVICE-NAME|ID"`

	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbServiceDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *nlbServiceDeleteCmd) CmdShort() string { return "Delete a Network Load Balancer service" }

func (c *nlbServiceDeleteCmd) CmdLong() string { return "" }

func (c *nlbServiceDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbServiceDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {

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

	for _, service := range nlb.Services {
		if service.ID.String() == c.Service || service.Name == c.Service {

			if !c.Force {
				if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete service %q?", service.ID)) {
					return nil
				}
			}

			op, err := client.DeleteLoadBalancerService(ctx, nlb.ID, service.ID)
			utils.DecorateAsyncOperation(fmt.Sprintf("Deleting service %q...", c.Service), func() {
				_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			})
			return err
		}
	}

	return errors.New("service not found")
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(nlbServiceCmd, &nlbServiceDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
