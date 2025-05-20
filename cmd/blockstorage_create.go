package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageCreateCmd struct {
	CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"create"`

	Name string `cli-arg:"#" cli-usage:"NAME"`

	Size     int64             `cli-usage:"block storage volume size (default: 10)"`
	Labels   map[string]string `cli-flag:"label" cli-usage:"block storage volume labels (format: key=value)"`
	Snapshot string            `cli-usage:"block storage volume snapshot NAME|ID"`
	Zone     v3.ZoneName       `cli-short:"z" cli-usage:"block storage zone"`
}

func (c *blockStorageCreateCmd) CmdAliases() []string { return GCreateAlias }

func (c *blockStorageCreateCmd) CmdShort() string { return "Create a Block Storage Volume" }

func (c *blockStorageCreateCmd) CmdLong() string {
	return fmt.Sprintf(`This command creates a Block Storage Volume.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&blockStorageShowOutput{}), ", "))
}

func (c *blockStorageCreateCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	CmdSetZoneFlagFromDefault(cmd)
	return CliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageCreateCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := GContext
	client, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	if c.Snapshot == "" && c.Size == 0 {
		c.Size = 10
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
		Name:                 c.Name,
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

	if !globalstate.Quiet {
		return (&blockStorageShowCmd{
			CliCommandSettings: c.CliCommandSettings,
			Name:               c.Name,
			Zone:               c.Zone,
		}).CmdRun(nil, nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(RegisterCLICommand(blockstorageCmd, &blockStorageCreateCmd{
		CliCommandSettings: DefaultCLICmdSettings(),
	}))
}
