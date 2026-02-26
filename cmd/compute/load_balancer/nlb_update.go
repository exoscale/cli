package load_balancer

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	v3 "github.com/exoscale/egoscale/v3"
)

type nlbUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"NAME|ID"`

	Description string            `cli-usage:"Network Load Balancer description"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"Network Load Balancer label (format: key=value)"`
	Name        string            `cli-usage:"Network Load Balancer name"`
	Zone        v3.ZoneName       `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbUpdateCmd) CmdAliases() []string { return nil }

func (c *nlbUpdateCmd) CmdShort() string { return "Update a Network Load Balancer" }

func (c *nlbUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates a Network Load Balancer.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbShowOutput{}), ", "),
	)
}

func (c *nlbUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exocmd.GContext

	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Description)) {
		nlbRequest.Description = c.Description
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Labels)) {
		nlbRequest.Labels = c.Labels
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Name)) {
		nlbRequest.Name = c.Name
		updated = true
	}

	if updated {

		op, err := client.UpdateLoadBalancer(ctx, n.ID, nlbRequest)
		if err != nil {
			return err
		}

		utils.DecorateAsyncOperation(
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
			CliCommandSettings:  c.CliCommandSettings,
			NetworkLoadBalancer: n.ID.String(),
			Zone:                c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(nlbCmd, &nlbUpdateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
