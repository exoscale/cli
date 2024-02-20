package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockstorageAttachCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"attach"`

	Volume   string      `cli-arg:"#" cli-usage:"NAME|ID"`
	Instance string      `cli-arg:"#" cli-usage:"NAME|ID"`
	Zone     v3.ZoneName `cli-short:"z" cli-usage:"block storage zone"`
}

func (c *blockstorageAttachCmd) cmdAliases() []string { return gCreateAlias }

func (c *blockstorageAttachCmd) cmdShort() string { return "Attach a Block Storage Volume" }

func (c *blockstorageAttachCmd) cmdLong() string {
	return fmt.Sprintf(`This command attaches Block Storage Volume.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockstorageShowOutput{}), ", "))
}

func (c *blockstorageAttachCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageAttachCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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
	if err != err {
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
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageAttachCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
