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
	egoscale "github.com/exoscale/egoscale/v2"
	exoapi "github.com/exoscale/egoscale/v2/api"
)

type nlbCreateCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	Description string            `cli-usage:"Network Load Balancer description"`
	Labels      map[string]string `cli-flag:"label" cli-usage:"Network Load Balancer label (format: key=value)"`
	Zone        string            `cli-short:"z" cli-usage:"Network Load Balancer zone"`
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
	nlb := &egoscale.NetworkLoadBalancer{
		Description: utils.NonEmptyStringPtr(c.Description),
		Labels: func() (v *map[string]string) {
			if len(c.Labels) > 0 {
				return &c.Labels
			}
			return
		}(),
		Name: &c.Name,
	}

	ctx := exoapi.WithEndpoint(exocmd.GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	var err error
	utils.DecorateAsyncOperation(fmt.Sprintf("Creating Network Load Balancer %q...", c.Name), func() {
		nlb, err = globalstate.EgoscaleClient.CreateNetworkLoadBalancer(ctx, c.Zone, nlb)
	})
	if err != nil {
		return err
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
	cobra.CheckErr(exocmd.RegisterCLICommand(nlbCmd, &nlbCreateCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
