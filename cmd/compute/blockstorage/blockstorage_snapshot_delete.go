package blockstorage

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

type blockStorageSnapshotDeleteCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name  string      `cli-arg:"#" cli-usage:"<NAME|ID>"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"block storage volume snapshot zone"`
	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *blockStorageSnapshotDeleteCmd) CmdAliases() []string { return exocmd.GDeleteAlias }

func (c *blockStorageSnapshotDeleteCmd) CmdShort() string {
	return "Delete a Block Storage Volume Snapshot"
}

func (c *blockStorageSnapshotDeleteCmd) CmdLong() string {
	return fmt.Sprintf(`This command deletes a Block Storage Volume Snapshot.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageShowOutput{}), ", "))
}

func (c *blockStorageSnapshotDeleteCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageSnapshotDeleteCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	resp, err := client.ListBlockStorageSnapshots(ctx)
	if err != nil {
		return err
	}
	snapshot, err := resp.FindBlockStorageSnapshot(c.Name)
	if err != nil {
		return err
	}

	if !c.Force {
		if !utils.AskQuestion(ctx, fmt.Sprintf("Are you sure you want to delete block storage volume snapshot %q?", c.Name)) {
			return nil
		}
	}

	op, err := client.DeleteBlockStorageSnapshot(ctx, snapshot.ID)
	if err != nil {
		return err
	}

	utils.DecorateAsyncOperation(fmt.Sprintf("Deleting block storage volume snapshot %q...", c.Name), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(blockstorageSnapshotCmd, &blockStorageSnapshotDeleteCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
