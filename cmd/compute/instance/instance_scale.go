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

type instanceScaleCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"scale"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`
	Type     string `cli-arg:"#" cli-usage:"SIZE"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceScaleCmd) CmdAliases() []string { return nil }

func (c *instanceScaleCmd) CmdShort() string { return "Scale a Compute instance" }

func (c *instanceScaleCmd) CmdLong() string {
	return fmt.Sprintf(`This commands scales a Compute instance to a different size.

Supported Compute instance type sizes: %s

Supported output template annotations: %s`,
		strings.Join(instanceTypeSizes, ", "),
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "))
}

func (c *instanceScaleCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceScaleCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	instance, err := instances.FindListInstancesResponseInstances(c.Instance)
	if err != nil {
		return err
	}
	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to scale instance %q?", c.Instance)) {
			return nil
		}
	}

	instanceTypes, err := client.ListInstanceTypes(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving instance type: %w", err)
	}
	instanceType, err := instanceTypes.FindInstanceTypeByIdOrFamilyAndSize(c.Type)
	if err != nil {
		return err
	}

	op, err := client.ScaleInstance(ctx, instance.ID, v3.ScaleInstanceRequest{
		InstanceType: &instanceType,
	})
	utils.DecorateAsyncOperation(fmt.Sprintf("Scaling instance %q...", c.Instance), func() {
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
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceScaleCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
