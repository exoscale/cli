package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageSnapshotUpdateCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Name   string            `cli-arg:"#" cli-usage:"NAME|ID"`
	Labels map[string]string `cli-flag:"label" cli-usage:"block storage volume snapshot label (format: key=value), clearing the labels is possible by passing [=]"`
	Zone   v3.ZoneName       `cli-short:"z" cli-usage:"block storage volume snapshot zone"`
	Rename string            `cli-usage:"rename block storage volume snapshot"`
}

func (c *blockStorageSnapshotUpdateCmd) CmdAliases() []string { return []string{"up"} }

func (c *blockStorageSnapshotUpdateCmd) CmdShort() string {
	return "Update a Block Storage Volume Snapshot"
}

func (c *blockStorageSnapshotUpdateCmd) CmdLong() string {
	return fmt.Sprintf(`This command updates a Block Storage Volume Snapshot.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageShowOutput{}), ", "))
}

func (c *blockStorageSnapshotUpdateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageSnapshotUpdateCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	snapshots, err := client.ListBlockStorageSnapshots(ctx)
	if err != nil {
		return err
	}

	snapshot, err := snapshots.FindBlockStorageSnapshot(c.Name)
	if err != nil {
		return err
	}

	var updated bool
	updateReq := v3.UpdateBlockStorageSnapshotRequest{}
	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Labels)) {
		updateReq.Labels = convertIfSpecialEmptyMap(c.Labels)

		updated = true
	}

	if cmd.Flags().Changed(MustCLICommandFlagName(c, &c.Rename)) {
		updateReq.Name = &c.Rename

		updated = true
	}

	if updated {
		op, err := client.UpdateBlockStorageSnapshot(ctx, snapshot.ID, updateReq)
		if err != nil {
			return err
		}

		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
		if err != nil {
			return err
		}
	}

	if updated && !globalstate.Quiet {
		name := c.Name
		if c.Rename != "" {
			name = c.Rename
		}
		return (&blockStorageSnapshotShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Name:               name,
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(blockstorageSnapshotCmd, &blockStorageSnapshotUpdateCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
