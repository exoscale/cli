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

type nlbCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	Description string            `cli-usage:"Network Load Balancer description"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"Network Load Balancer label (format: key=value)"`
	Zone        v3.ZoneName       `cli-short:"z" cli-usage:"Network Load Balancer zone"`
}

func (c *nlbCreateCmd) CmdAliases() []string { return exocmd.GCreateAlias }

func (c *nlbCreateCmd) CmdShort() string { return "Create a Network Load Balancer" }

func (c *nlbCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Network Load Balancer.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&nlbShowOutput{}), ", "))
}

func (c *nlbCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *nlbCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	nlb := v3.CreateLoadBalancerRequest{
		Description: c.Description,
		Labels:      c.Labels,
		Name:        c.Name,
	}
	var err error

	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	op, err := client.CreateLoadBalancer(ctx, nlb)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Creating Network Load Balancer %q...", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&nlbShowCmd{
			CliCommandSettings:  c.CliCommandSettings,
			NetworkLoadBalancer: op.Reference.ID.String(),
			Zone:                c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(nlbCmd, &nlbCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
