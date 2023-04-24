package cmd

import (
	"fmt"
	"strings"

	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
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
		strings.Join(output.output.OutputterTemplateAnnotations(&nlbShowOutput{}), ", "),
	)
}

func (c *nlbUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	var updated bool

	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	nlb, err := cs.FindNetworkLoadBalancer(ctx, c.Zone, c.NetworkLoadBalancer)
	if err != nil {
		return err
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Description)) {
		nlb.Description = &c.Description
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Labels)) {
		nlb.Labels = &c.Labels
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		nlb.Name = &c.Name
		updated = true
	}

	if updated {
		decorateAsyncOperation(
			fmt.Sprintf("Updating Network Load Balancer %q...", c.NetworkLoadBalancer),
			func() {
				if err = cs.UpdateNetworkLoadBalancer(ctx, c.Zone, nlb); err != nil {
					return
				}
			})
		if err != nil {
			return err
		}
	}

	if !gQuiet {
		return (&nlbShowCmd{
			cliCommandSettings:  c.cliCommandSettings,
			NetworkLoadBalancer: *nlb.ID,
			Zone:                c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbCmd, &nlbUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))

	// FIXME: remove this someday.
	cobra.CheckErr(registerCLICommand(deprecatedNLBCmd, &nlbUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
