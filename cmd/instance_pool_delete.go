package cmd

import (
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instancePoolDeleteCmd struct {
	_ bool `cli-cmd:"delete"`

	InstancePool string `cli-arg:"#" cli-usage:"NAME|ID"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"Instance Pool zone"`
}

func (c *instancePoolDeleteCmd) cmdAliases() []string { return gRemoveAlias }

func (c *instancePoolDeleteCmd) cmdShort() string { return "Delete an Instance Pool" }

func (c *instancePoolDeleteCmd) cmdLong() string { return "" }

func (c *instancePoolDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instancePoolDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instancePool, err := cs.FindInstancePool(ctx, c.Zone, c.InstancePool)
	if err != nil {
		return err
	}

	// Ensure the Instance Pool is not attached to an NLB service.
	nlbs, err := cs.ListNetworkLoadBalancers(gContext, c.Zone)
	if err != nil {
		return fmt.Errorf("unable to list Network Load Balancers: %v", err)
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete Instance Pool %q?", c.InstancePool)) {
			return nil
		}
	}

	for _, nlb := range nlbs {
		for _, svc := range nlb.Services {
			if svc.InstancePoolID == instancePool.ID {
				return fmt.Errorf(
					"Instance Pool %q is still referenced by NLB service %s/%s", // nolint:golint
					*instancePool.Name,
					*nlb.Name,
					*svc.Name,
				)
			}
		}
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting Instance Pool %q...", c.InstancePool), func() {
		err = cs.DeleteInstancePool(ctx, c.Zone, *instancePool.ID)
	})
	if err != nil {
		return err
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instancePoolCmd, &instancePoolDeleteCmd{}))
}
