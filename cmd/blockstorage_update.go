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
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Name   string            `cli-arg:"#" cli-usage:"NAME|ID"`
	Size   int64             `cli-usage:"block storage volume size"`
	Labels map[string]string `cli-flag:"label" cli-usage:"block storage volume label (format: key=value), clearing the labels is possible by passing [=]"`
	Zone   v3.ZoneName       `cli-short:"z" cli-usage:"block storage volume zone"`
	Rename string            `cli-usage:"rename block storage volume"`
}

func (c *blockStorageUpdateCmd) CmdAliases() []string { return []string{"up"} }

func (c *blockStorageUpdateCmd) CmdShort() string { return "Update a Block Storage Volume" }

func (c *blockStorageUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates a Block Storage Volume.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageShowOutput{}), ", "))
}

func (c *blockStorageUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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
	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Labels)) {
		updateReq.Labels = convertIfSpecialEmptyMap(c.Labels)

		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Rename)) {
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
			CliCommandSettings: c.CliCommandSettings,
			Name:               name,
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(blockstorageCmd, &blockStorageUpdateCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
