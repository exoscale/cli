package instance

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

type instanceEIPDetachCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

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
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "),
	)
}

func (c *instanceEIPDetachCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceEIPDetachCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	instancesList, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	instance, err := findInstance(instancesList, c.Instance, c.Zone)
	if err != nil {
		return fmt.Errorf("error retrieving Instance: %w", err)
	}

	elasticIPs, err := client.ListElasticIPS(ctx)
	if err != nil {
		return err
	}
	elasticIP, err := elasticIPs.FindElasticIP(c.ElasticIP)
	if err != nil {
		return fmt.Errorf("error retrieving Elastic IP: %w", err)
	}

	op, err := client.DetachInstanceFromElasticIP(ctx, elasticIP.ID, v3.DetachInstanceFromElasticIPRequest{
		Instance: &v3.InstanceTarget{
			ID: instance.ID,
		},
	})
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf(
		"Detaching instance %q from Elastic IP %q...",
		c.Instance,
		c.ElasticIP,
	), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&instanceShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Instance:           instance.ID.String(),
			Zone:               v3.ZoneName(c.Zone),
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceEIPCmd, &instanceEIPDetachCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
