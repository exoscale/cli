package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageSnapshotUpdateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"update"`

	Name   string            `cli-arg:"#" cli-usage:"NAME|ID"`
	Labels map[string]string `cli-flag:"label" cli-usage:"block storage volume snapshot label (format: key=value)"`
	Zone   v3.ZoneName       `cli-short:"z" cli-usage:"block storage volume snapshot zone"`
	Rename string            `cli-usage:"rename block storage volume snapshot"`
}

func (c *blockStorageSnapshotUpdateCmd) cmdAliases() []string { return []string{"up"} }

func (c *blockStorageSnapshotUpdateCmd) cmdShort() string {
	return "Update a Block Storage Volume Snapshot"
}

func (c *blockStorageSnapshotUpdateCmd) cmdLong() string {
	return fmt.Sprintf(`This command updates a Block Storage Volume Snapshot.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageShowOutput{}), ", "))
}

func (c *blockStorageSnapshotUpdateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageSnapshotUpdateCmd) cmdRun(cmd *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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
	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Labels)) {
		updateReq.Labels = convertIfSpecialEmptyMap(c.Labels)

		updated = true
	}

	if cmd.Flags().Changed(mustCLICommandFlagName(c, &c.Rename)) {
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

		time.Sleep(1 * time.Second)
	}

	if updated && !globalstate.Quiet {
		name := c.Name
		if c.Rename != "" {
			name = c.Rename
		}
		return (&blockStorageSnapshotShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Name:               name,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageSnapshotCmd, &blockStorageSnapshotUpdateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
