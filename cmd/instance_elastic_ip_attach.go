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

type instanceEIPAttachCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"attach"`

	Instance  string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	ElasticIP string `cli-arg:"#" cli-usage:"ELASTIC-IP-ADDRESS|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceEIPAttachCmd) CmdAliases() []string { return nil }

func (c *instanceEIPAttachCmd) CmdShort() string {
	return "Attach an Elastic IP to a Compute instance"
}

func (c *instanceEIPAttachCmd) CmdLong() string {
	return fmt.Sprintf(`This command attaches an Elastic IP address to a Compute instance.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instanceEIPAttachCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceEIPAttachCmd) CmdRun(_ *cobra.Command, _ []string) error {
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
		"Attaching Elastic IP %q to instance %q...",
		c.ElasticIP,
		c.Instance,
	), func() {
		if err = globalstate.EgoscaleClient.AttachInstanceToElasticIP(
			ctx,
			c.Zone,
			instance,
			elasticIP,
		); err != nil {
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
	cobra.CheckErr(RegisterCLICommand(instanceEIPCmd, &instanceEIPAttachCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
