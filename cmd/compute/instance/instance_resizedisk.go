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

type instanceResizeDiskCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"resize-disk"`

	Instance string `cli-arg:"#" cli-usage:"NAME|ID"`
	Size     int64  `cli-arg:"#"`

	Force bool   `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone  string `cli-short:"z" cli-usage:"instance zone"`
}

func (c *instanceResizeDiskCmd) CmdAliases() []string { return nil }

func (c *instanceResizeDiskCmd) CmdShort() string { return "Resize a Compute instance disk" }

func (c *instanceResizeDiskCmd) CmdLong() string {
	return fmt.Sprintf(`This commands grows a Compute instance's disk to a larger size.'

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&InstanceShowOutput{}), ", "))
}

func (c *instanceResizeDiskCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceResizeDiskCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	instances, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	instance, err := findInstance(instances, c.Instance, c.Zone)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to resize the disk of instance %q?", c.Instance)) {
			return nil
		}
	}

	op, err := client.ResizeInstanceDisk(ctx, instance.ID, v3.ResizeInstanceDiskRequest{
		DiskSize: c.Size,
	})
	if err != nil {
		return err
	}
	utils.DecorateAsyncOperation(fmt.Sprintf("Resizing disk of instance %q...", c.Instance), func() {
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
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceCmd, &instanceResizeDiskCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
