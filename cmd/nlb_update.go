package cmd

import (
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

var nlbResetFields = []string{
	"labels",
}

type nlbUpdateCmd struct {
	_ bool `cli-cmd:"update"`

	NetworkLoadBalancer string `cli-arg:"#" cli-usage:"NAME|ID"`

	Description string            `cli-usage:"Network Load Balancer description"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"Network Load Balancer label (format: key=value)"`
	Name        string            `cli-usage:"Network Load Balancer name"`
	ResetFields []string          `cli-flag:"reset" cli-usage:"properties to reset to default value"`
	Zone        string            `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbUpdateCmd) cmdAliases() []string { return nil }

func (c *nlbUpdateCmd) cmdShort() string { return "Update a Network Load Balancer" }

func (c *nlbUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates a Network Load Balancer.

Supported output template annotations: %s

Support values for --reset flag: %s`,
		strings.Join(outputterTemplateAnnotations(&nlbShowOutput{}), ", "),
		strings.Join(nlbResetFields, ", "),
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
		nlb.Description = c.Description
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Labels)) {
		nlb.Labels = c.Labels
		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Name)) {
		nlb.Name = c.Name
		updated = true
	}

	decorateAsyncOperation(fmt.Sprintf("Updating Network Load Balancer %q...", nlb.Name), func() {
		if updated {
			if err = cs.UpdateNetworkLoadBalancer(ctx, c.Zone, nlb); err != nil {
				return
			}
		}

		for _, f := range c.ResetFields {
			switch f {
			case "labels":
				err = nlb.ResetField(ctx, &nlb.Labels)
			}
			if err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return output(showNLB(c.Zone, nlb.ID))
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbCmd, &nlbUpdateCmd{}))
}
