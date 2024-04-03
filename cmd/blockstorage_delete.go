package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name  string      `cli-arg:"#" cli-usage:"NAME|ID"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"block storage volume zone"`
	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *blockStorageDeleteCmd) cmdAliases() []string { return gDeleteAlias }

func (c *blockStorageDeleteCmd) cmdShort() string { return "Delete a Block Storage Volume" }

func (c *blockStorageDeleteCmd) cmdLong() string {
	return fmt.Sprintf(`This command deletes a Block Storage Volume.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageShowOutput{}), ", "))
}

func (c *blockStorageDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListBlockStorageVolumes(ctx)
	if err != nil {
		return err
	}
	volume, err := resp.FindBlockStorageVolume(c.Name)
	if err != nil {
		return err
	}

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete block storage volume %q?", c.Name)) {
			return nil
		}
	}

	op, err := client.DeleteBlockStorageVolume(ctx, volume.ID)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting block storage volume %q...", c.Name), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockStorageDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
