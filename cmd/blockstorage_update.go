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

	Name string `cli-arg:"#" cli-usage:"NAME|ID"`
	Size int64  `cli-usage:"block storage volume size"`
	// TODO(pej): Re-enable it when API is up to date on this call.
	// Labels map[string]string `cli-flag:"label" cli-usage:"block storage volume label (format: key=value)"`
	Zone v3.ZoneName `cli-short:"z" cli-usage:"block storage volume zone"`
}

func (c *blockStorageUpdateCmd) cmdAliases() []string { return gCreateAlias }

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

func (c *blockStorageUpdateCmd) cmdRun(_ *cobra.Command, _ []string) error {
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

	if c.Size == 0 {
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
	})
	if err != nil {
		return err
	}

	if updated && !globalstate.Quiet {
		return (&blockStorageShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Name:               c.Name,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockStorageUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
