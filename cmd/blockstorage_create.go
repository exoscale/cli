package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockstorageCreateCmd struct {
	cliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	Size     int64             `cli-usage:"block storage volume size"`
	Snapshot string            `cli-usage:"block storage volume snapshot NAME|ID"`
	Labels   map[string]string `cli-flag:"label" cli-usage:"block storage volume label (format: key=value)"`
	Zone     v3.ZoneName       `cli-short:"z" cli-usage:"block storage zone"`
}

func (c *blockstorageCreateCmd) cmdAliases() []string { return gCreateAlias }

func (c *blockstorageCreateCmd) cmdShort() string { return "Create a Block Storage Volume" }

func (c *blockstorageCreateCmd) cmdLong() string {
	return fmt.Sprintf(`This command creates Block Storage Volume.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockstorageShowOutput{}), ", "))
}

func (c *blockstorageCreateCmd) cmdPreRun(cmd *cobra.Command, args []string) error {
	cmdSetZoneFlagFromDefault(cmd)
	return cliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockstorageCreateCmd) cmdRun(_ *cobra.Command, _ []string) error {
	ctx := gContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	var snapshot *v3.BlockStorageSnapshotTarget
	if c.Snapshot != "" {
		snapshots, err := client.ListBlockStorageSnapshots(ctx)
		if err != nil {
			return err
		}
		s, err := snapshots.FindBlockStorageSnapshot(c.Snapshot)
		if err != nil {
			return err
		}
		snapshot = &v3.BlockStorageSnapshotTarget{ID: s.ID}
	}
	req := v3.CreateBlockStorageVolumeRequest{
		Name:                 "",
		Size:                 c.Size,
		Labels:               c.Labels,
		BlockStorageSnapshot: snapshot,
	}

	if err := client.Validate(req); err != nil {
		return err
	}

	op, err := client.CreateBlockStorageVolume(ctx, req)
	if err != nil {
		return err
	}

	decorateAsyncOperation(fmt.Sprintf("Creating block storage volume %q...", c.Name), func() {
		op, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	bs, err := client.GetBlockStorageVolume(ctx, op.Reference.ID)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return (&blockstorageShowCmd{
			cliCommandSettings: c.cliCommandSettings,
			Name:               bs.ID.String(),
		}).cmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageCmd, &blockstorageCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
