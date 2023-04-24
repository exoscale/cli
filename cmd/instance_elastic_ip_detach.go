package cmd

import (
	"errors"
	"fmt"
	"strings"

	exoapi "github.com/exoscale/egoscale/v2/api"
	"github.com/spf13/cobra"
)

type instanceEIPDetachCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"detach"`

	Instance  string `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	ElasticIP string `cli-arg:"#" cli-usage:"ELASTIC-IP-ADDRESS|ID"`

	Zone string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceEIPDetachCmd) cmdAliases() []string { return nil }

func (c *instanceEIPDetachCmd) cmdShort() string {
	return "Detach a Compute instance from a Elastic IP"
}

func (c *instanceEIPDetachCmd) cmdLong() string {
	return fmt.Sprintf(`This command detaches an Elastic IP address from a Compute instance.

Supported output template annotations: %s`,
		strings.Join(output.OutputterTemplateAnnotations(&instanceShowOutput{}), ", "),
	)
}

func (c *instanceEIPDetachCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceEIPDetachCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := exoapi.WithEndpoint(gContext, exoapi.NewReqEndpoint(gCurrentAccount.Environment, c.Zone))

	instance, err := cs.FindInstance(ctx, c.Zone, c.Instance)
	if err != nil {
		if errors.Is(err, exoapi.ErrNotFound) {
			return fmt.Errorf("resource not found in zone %q", c.Zone)
		}
		return err
	}

	elasticIP, err := cs.FindElasticIP(ctx, c.Zone, c.ElasticIP)
	if err != nil {
		return fmt.Errorf("error retrieving Elastic IP: %w", err)
	}

	decorateAsyncOperation(fmt.Sprintf(
		"Detaching instance %q from Elastic IP %q...",
		c.Instance,
		c.ElasticIP,
	), func() {
		if err = cs.DetachInstanceFromElasticIP(ctx, c.Zone, instance, elasticIP); err != nil {
			return
		}
	})
	if err != nil {
		return err
	}

	if !gQuiet {
		return (&instanceShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Instance:           *instance.ID,
			Zone:               c.Zone,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(instanceEIPCmd, &instanceEIPDetachCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
