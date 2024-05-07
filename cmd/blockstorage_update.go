package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Name   string            `cli-arg:"#" cli-usage:"NAME|ID"`
	Size   int64             `cli-usage:"block storage volume size"`
	Labels map[string]string `cli-flag:"label" cli-usage:"block storage volume label (format: key=value)"`
	Zone   v3.ZoneName       `cli-short:"z" cli-usage:"block storage volume zone"`
	Rename string            `cli-usage:"rename block storage volume"`
}

func (c *blockStorageUpdateCmd) cmdAliases() []string { return []string{"up"} }

func (c *blockStorageUpdateCmd) cmdShort() string { return "Update a Block Storage Volume" }

func (c *blockStorageUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates a Block Storage Volume.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageShowOutput{}), ", "))
}

func (c *blockStorageUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
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

	var resized bool

	if c.Size > 0 {
		decorateAsyncOperation(fmt.Sprintf("Updating block storage volume %q...", c.Name), func() {
			_, err = client.ResizeBlockStorageVolume(ctx, volume.ID,
				v3.ResizeBlockStorageVolumeRequest{
					Size: c.Size,
				},
			)
			if err != nil {
				return
			}
			resized = true
		})
		if err != nil {
			return err
		}
	}

	var updated bool
	updateReq := v3.UpdateBlockStorageVolumeRequest{}
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Labels)) {
		updateReq.Labels = c.Labels

		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Rename)) {
		updateReq.Name = &c.Rename

		updated = true
	}

	if updated {
		op, err := client.UpdateBlockStorageVolume(ctx, volume.ID, updateReq)
		if err != nil {
			return err
		}

		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return err
		}
	}

	if (resized || updated) && !globalstate.Quiet {
		name := c.Name
		if c.Rename != "" {
			name = c.Rename
		}
		return (&blockStorageShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Name:               name,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockStorageUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
