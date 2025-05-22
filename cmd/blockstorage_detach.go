package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageDetachCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"detach"`

	Volume string      `cli-arg:"#" cli-usage:"NAME|ID"`
	Force  bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
	Zone   v3.ZoneName `cli-short:"z" cli-usage:"block storage volume zone"`
}

func (c *blockStorageDetachCmd) CmdAliases() []string { return []string{"d"} }

func (c *blockStorageDetachCmd) CmdShort() string { return "Detach a Block Storage Volume" }

func (c *blockStorageDetachCmd) CmdLong() string {
	return fmt.Sprintf(`This command detaches Block Storage Volume.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageShowOutput{}), ", "))
}

func (c *blockStorageDetachCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageDetachCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	if !c.Force {
		if !askQuestion(fmt.Sprintf("Are you sure you want to detach block storage volume %q?", c.Volume)) {
			return nil
		}
	}

	op, err := client.DetachBlockStorageVolume(ctx, volume.ID)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Detaching block storage volume %q...", c.Volume), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(RegisterCLICommand(blockstorageCmd, &blockStorageDetachCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
