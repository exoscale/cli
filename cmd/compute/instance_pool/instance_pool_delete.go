package instance_pool

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type instancePoolDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	InstancePool string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolDeleteCmd) CmdAliases() []string { return exocmd.GRemoveAlias }

func (c *instancePoolDeleteCmd) CmdShort() string { return "Delete an Instance Pool" }

func (c *instancePoolDeleteCmd) CmdLong() string { return "" }

func (c *instancePoolDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instancePool, err := globalstate.EgoscaleClient.FindInstancePool(ctx, c.Zone, c.InstancePool)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	// Ensure the Instance Pool is not attached to an NLB service.
	nlbs, err := globalstate.EgoscaleClient.ListNetworkLoadBalancers(ctx, c.Zone)
	if err != nil {
		return fmt.Errorf("unable to list Network Load Balancers: %v", err)
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete Instance Pool %q?", c.InstancePool)) {
			return nil
		}
	}

	for _, nlb := range nlbs {
		for _, svc := range nlb.Services {
			if svc.InstancePoolID == instancePool.ID {
				return fmt.Errorf(
					"instance Pool %q is still referenced by NLB service %s/%s",
					*instancePool.Name,
					*nlb.Name,
					*svc.Name,
				)
			}
		}
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting Instance Pool %q...", c.InstancePool), func() {
		err = globalstate.EgoscaleClient.DeleteInstancePool(ctx, c.Zone, instancePool)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instancePoolCmd, &instancePoolDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
