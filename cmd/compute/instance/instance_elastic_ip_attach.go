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

type instanceEIPAttachCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

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
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "),
	)
}

func (c *instanceEIPAttachCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceEIPAttachCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	instancesList, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	instance, err := instancesList.FindListInstancesResponseInstances(c.Instance)
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

	op, err := client.AttachInstanceToElasticIP(ctx, elasticIP.ID, v3.AttachInstanceToElasticIPRequest{
		Instance: &v3.InstanceTarget{
			ID: instance.ID,
		},
	})
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf(
		"Attaching Elastic IP %q to instance %q...",
		c.ElasticIP,
		c.Instance,
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
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceEIPCmd, &instanceEIPAttachCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
