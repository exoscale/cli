package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageAttachCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"attach"`

	Volume   string      `cli-arg:"#" cli-usage:"NAME|ID"`
	Instance string      `cli-arg:"#" cli-usage:"INSTANCE-NAME|ID"`
	Zone     v3.ZoneName `cli-short:"z" cli-usage:"block storage zone"`
}

func (c *blockStorageAttachCmd) CmdAliases() []string { return []string{"a"} }

func (c *blockStorageAttachCmd) CmdShort() string { return "Attach a Block Storage Volume" }

func (c *blockStorageAttachCmd) CmdLong() string {
	return fmt.Sprintf(`This command attaches a Block Storage Volume to a Compute Instance.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageShowOutput{}), ", "))
}

func (c *blockStorageAttachCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageAttachCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	volumes, err := client.ListBlockStorageVolumes(ctx)
	if err != nil {
		return err
	}

	volume, err := volumes.FindBlockStorageVolume(c.Volume)
	if err != nil {
		return err
	}

	resp, err := client.ListInstances(ctx)
	if err != nil {
		return err
	}
	instance, err := resp.FindListInstancesResponseInstances(c.Instance)
	if err != nil {
		return err
	}

	op, err := client.AttachBlockStorageVolumeToInstance(ctx, volume.ID,
		v3.AttachBlockStorageVolumeToInstanceRequest{
			Instance: &v3.InstanceTarget{
				ID: instance.ID,
			},
		},
	)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Attaching volume %q to instance %q...", c.Volume, c.Instance), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(RegisterCLICommand(blockstorageCmd, &blockStorageAttachCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
