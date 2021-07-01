package cmd

import (
	"errors"
	"fmt"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type nlbServiceDeleteCmd struct {
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
	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete service %q?", c.Service)) {
			return nil
		}
	}

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	nlb, err := cs.FindNetworkLoadBalancer(ctx, c.Zone, c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	for _, s := range nlb.Services {
		if *s.ID == c.Service || *s.Name == c.Service {
			s := s
			decorateAsyncOperation(fmt.Sprintf("Deleting service %q...", c.Service), func() {
				err = nlb.DeleteService(ctx, s)
			})
			if err != nil {
				return err
			}

			return nil
		}
	}

	return errors.New("service not found")
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbServiceCmd, &nlbServiceDeleteCmd{}))
}
