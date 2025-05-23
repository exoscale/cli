package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type nlbCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	Description string            `cli-usage:"Network Load Balancer description"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"Network Load Balancer label (format: key=value)"`
	Zone        v3.ZoneName       `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *nlbCreateCmd) cmdShort() string { return "Create a Network Load Balancer" }

func (c *nlbCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates a Network Load Balancer.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbShowOutput{}), ", "))
}

func (c *nlbCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	nlb := v3.CreateLoadBalancerRequest{
		Description: c.Description,
		Labels:      c.Labels,
		Name:        c.Name,
	}
	var err error

	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	op, err := client.CreateLoadBalancer(ctx, nlb)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating Network Load Balancer %q...", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&nlbShowCmd{
			cliCommandSettings:  c.cliCommandSettings,
			NetworkLoadBalancer: op.Reference.ID.String(),
			Zone:                c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(nlbCmd, &nlbCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
