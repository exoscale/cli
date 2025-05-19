package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	v3 "github.com/exoscale/egoscale/v3"
)

type nlbServiceDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"LOAD-BALANCER-NAME|ID"`
	Service             string `cli-arg:"#" cli-usage:"SERVICE-NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbServiceDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *nlbServiceDeleteCmd) cmdShort() string { return "Delete a Network Load Balancer service" }

func (c *nlbServiceDeleteCmd) cmdLong() string { return "" }

func (c *nlbServiceDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbServiceDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {

	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
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
				if !askQuestion(fmt.Sprintf("Are you sure you want to delete service %q?", service.ID)) {
					return nil
				}
			}

			op, err := client.DeleteLoadBalancerService(ctx, nlb.ID, service.ID)
			decorateAsyncOperation(fmt.Sprintf("Deleting service %q...", c.Service), func() {
				_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			})
			return err
		}
	}

	return errors.New("service not found")
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbServiceCmd, &nlbServiceDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
