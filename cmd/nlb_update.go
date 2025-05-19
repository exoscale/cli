package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type nlbUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"NAME|ID"`

	Description string            `cli-usage:"Network Load Balancer description"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"Network Load Balancer label (format: key=value)"`
	Name        string            `cli-usage:"Network Load Balancer name"`
	Zone        string            `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbUpdateCmd) cmdAliases() []string { return nil }

func (c *nlbUpdateCmd) cmdShort() string { return "Update a Network Load Balancer" }

func (c *nlbUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates a Network Load Balancer.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbShowOutput{}), ", "),
	)
}

func (c *nlbUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := gContext

	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}
	nlbs, err := client.ListLoadBalancers(ctx)
	if err != nil {
		return err
	}

	n, err := nlbs.FindLoadBalancer(c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	nlbRequest := v3.UpdateLoadBalancerRequest{}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		nlbRequest.Description = c.Description
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Labels)) {
		nlbRequest.Labels = c.Labels
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		nlbRequest.Name = c.Name
		updated = true
	}

	if updated {

		op, err := client.UpdateLoadBalancer(ctx, n.ID, nlbRequest)

		decorateAsyncOperation(
			fmt.Sprintf("Updating Network Load Balancer %q...", c.NetworkLoadBalancer),
			func() {
				_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			})
		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&nlbShowCmd{
			cliCommandSettings:  c.cliCommandSettings,
			NetworkLoadBalancer: n.ID.String(),
			Zone:                c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbCmd, &nlbUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
