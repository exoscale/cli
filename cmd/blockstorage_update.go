package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockstorageUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Name   string            `cli-arg:"#" cli-usage:"NAME|ID"`
	Size   int64             `cli-usage:"block storage volume size"`
	Labels map[string]string `cli-flag:"label" cli-usage:"block storage volume label (format: key=value)"`
	Zone   v3.ZoneName       `cli-short:"z" cli-usage:"block storage zone"`
}

func (c *blockstorageUpdateCmd) cmdAliases() []string { return gCreateAlias }

func (c *blockstorageUpdateCmd) cmdShort() string { return "Update a Block Storage Volume" }

func (c *blockstorageUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command Updates Block Storage Volume.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockstorageShowOutput{}), ", "))
}

func (c *blockstorageUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageUpdateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	volumes, err := client.ListBlockStorageVolumes(ctx)
	if err != nil {
		return err
	}

	volume, err := volumes.FindBlockStorageVolume(c.Name)
	if err != nil {
		return err
	}

	if len(c.Labels) == 0 && c.Size == 0 {
		return nil
	}

	var updated bool
	decorateAsyncOperation(fmt.Sprintf("Updating block storage volume %q...", c.Name), func() {
		if c.Size != 0 {
			_, err = client.ResizeBlockStorageVolume(ctx, volume.ID,
				v3.ResizeBlockStorageVolumeRequest{
					Size: c.Size,
				},
			)
			if err != nil {
				return
			}
			updated = true
		}

		if len(c.Labels) > 0 {
			for k, v := range c.Labels {
				volume.Labels[k] = v
			}

			var op *v3.Operation
			op, err = client.UpdateBlockStorageVolumeLabels(ctx, volume.ID,
				v3.UpdateBlockStorageVolumeLabelsRequest{
					Labels: volume.Labels,
				},
			)
			if err != nil {
				return
			}

			_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
			updated = true
		}
	})
	if err != nil {
		return err
	}

	if updated && !globalstate.Quiet {
		return (&blockstorageShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Name:               c.Name,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
