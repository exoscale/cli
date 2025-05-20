package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	"github.com/exoscale/cli/utils"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type nlbUpdateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"NAME|ID"`

	Description string            `cli-usage:"Network Load Balancer description"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"Network Load Balancer label (format: key=value)"`
	Name        string            `cli-usage:"Network Load Balancer name"`
	Zone        string            `cli-short:"z" cli-usage:"Network Load Balancer zone"`
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

	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	nlb, err := globalstate.EgoscaleClient.FindNetworkLoadBalancer(ctx, c.Zone, c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Description)) {
		nlb.Description = &c.Description
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Labels)) {
		nlb.Labels = &c.Labels
		updated = true
	}

	if cmd.Flags().Changed(exocmd.MustCLICommandFlagName(c, &c.Name)) {
		nlb.Name = &c.Name
		updated = true
	}

	if updated {
		utils.DecorateAsyncOperation(
			fmt.Sprintf("Updating Network Load Balancer %q...", c.NetworkLoadBalancer),
			func() {
				if err = globalstate.EgoscaleClient.UpdateNetworkLoadBalancer(ctx, c.Zone, nlb); err != nil {
					return
				}
			})
		if err != nil {
			return err
		}
	}

	if !globalstate.Quiet {
		return (&nlbShowCmd{
			CliCommandSettings:  c.CliCommandSettings,
			NetworkLoadBalancer: *nlb.ID,
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
