package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageSnapshotDeleteCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"delete"`

	Name  string      `cli-arg:"#" cli-usage:"<NAME|ID>"`
	Zone  v3.ZoneName `cli-short:"z" cli-usage:"block storage volume snapshot zone"`
	Force bool        `cli-short:"f" cli-usage:"don't prompt for confirmation"`
}

func (c *blockStorageSnapshotDeleteCmd) cmdAliases() []string { return gDeleteAlias }

func (c *blockStorageSnapshotDeleteCmd) cmdShort() string {
	return "Delete a Block Storage Volume Snapshot"
}

func (c *blockStorageSnapshotDeleteCmd) cmdLong() string {
	return fmt.Sprintf(`This command deletes a Block Storage Volume Snapshot.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageShowOutput{}), ", "))
}

func (c *blockStorageSnapshotDeleteCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageSnapshotDeleteCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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
		if !askQuestion(fmt.Sprintf("Are you sure you want to delete block storage volume snapshot %q?", c.Name)) {
			return nil
		}
	}

	op, err := client.DeleteBlockStorageSnapshot(ctx, snapshot.ID)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Deleting block storage volume snapshot %q...", c.Name), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})

	return err
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageSnapshotCmd, &blockStorageSnapshotDeleteCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
