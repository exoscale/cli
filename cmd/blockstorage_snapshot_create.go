package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockstorageSnapshotCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Volume string            `cli-arg:"#" cli-usage:"<volume NAME|ID>"`
	Name   string            `cli-flag:"name" cli-usage:"block storage volume snapshot name"`
	Labels map[string]string `cli-flag:"label" cli-usage:"block storage volume snapshot label (format: key=value)"`
	Zone   v3.ZoneName       `cli-short:"z" cli-usage:"block storage volume snapshot zone"`
}

func (c *blockstorageSnapshotCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *blockstorageSnapshotCreateCmd) cmdShort() string {
	return "Create a Block Storage Volume Snapshot"
}

func (c *blockstorageSnapshotCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates Block Storage Volume Snapshot.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockstorageShowOutput{}), ", "))
}

func (c *blockstorageSnapshotCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageSnapshotCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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
			Labels: c.Labels,
			Name:   c.Name,
		},
	)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Snapshoting block storage volume %q...", c.Volume), func() {
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
		return (&blockstorageSnapshotShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Name:               name,
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageSnapshotCmd, &blockstorageSnapshotCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
