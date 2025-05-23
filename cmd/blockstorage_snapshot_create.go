package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageSnapshotCreateCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Volume string            `cli-arg:"#" cli-usage:"<volume NAME|ID>"`
	Name   string            `cli-flag:"name" cli-usage:"block storage volume snapshot name"`
	Labels map[string]string `cli-flag:"label" cli-usage:"block storage volume snapshot labels (format: key=value)"`
	Zone   v3.ZoneName       `cli-short:"z" cli-usage:"block storage volume snapshot zone"`
}

func (c *blockStorageSnapshotCreateCmd) CmdAliases() []string { return GCreateAlias }

func (c *blockStorageSnapshotCreateCmd) CmdShort() string {
	return "Create a Block Storage Volume Snapshot"
}

func (c *blockStorageSnapshotCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Block Storage Volume Snapshot.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageShowOutput{}), ", "))
}

func (c *blockStorageSnapshotCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageSnapshotCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
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

	op, err := client.CreateBlockStorageSnapshot(ctx, volume.ID,
		v3.CreateBlockStorageSnapshotRequest{
			Name:   c.Name,
			Labels: c.Labels,
		},
	)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Snapshotting block storage volume %q...", c.Volume), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	name := c.Name
	if c.Name == "" {
		name = op.Reference.ID.String()
	}

	if !globalstate.Quiet {
		return (&blockStorageSnapshotShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Name:               name,
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(blockstorageSnapshotCmd, &blockStorageSnapshotCreateCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
