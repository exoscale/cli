package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/account"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	exoapi "github.com/exoscale/egoscale/v2/api"
	v3 "github.com/exoscale/egoscale/v3"
)

type instanceEIPDetachCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"detach"`

	Instance  string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	ElasticIP string `cli-arg:"#" cli-usage:"ELASTIC-IP-ADDRESS|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceEIPDetachCmd) CmdAliases() []string { return nil }

func (c *instanceEIPDetachCmd) CmdShort() string {
	return "Detach a Compute instance from a Elastic IP"
}

func (c *instanceEIPDetachCmd) CmdLong() string {
	return fmt.Sprintf(`This command detaches an Elastic IP address from a Compute instance.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instanceEIPDetachCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceEIPDetachCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(GContext, exoapi.NewReqEndpoint(account.CurrentAccount.Environment, c.Zone))

	instance, err := globalstate.EgoscaleClient.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	elasticIP, err := globalstate.EgoscaleClient.FindElasticIP(ctx, c.Zone, c.ElasticIP)
	if err != nil {
		return fmt.Errorf("error retrieving Elastic IP: %w", err)
	}

	decorateAsyncOperation(fmt.Sprintf(
		"Detaching instance %q from Elastic IP %q...",
		c.Instance,
		c.ElasticIP,
	), func() {
		if err = globalstate.EgoscaleClient.DetachInstanceFromElasticIP(ctx, c.Zone, instance, elasticIP); err != nil {
			return
		}
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           *instance.ID,
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(instanceEIPCmd, &instanceEIPDetachCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
