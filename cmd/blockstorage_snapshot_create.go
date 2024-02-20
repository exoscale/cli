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

	Volume string `cli-arg:"#" cli-usage:"NAME|ID"`

	Size     int64             `cli-usage:"block storage volume size"`
	Snapshot string            `cli-usage:"block storage volume snapshot NAME|ID"`
	Labels   map[string]string `cli-flag:"label" cli-usage:"block storage volume label (format: key=value)"`
	Zone     v3.ZoneName       `cli-short:"z" cli-usage:"block storage zone"`
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
	_, err := switchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
	if err != nil {
		return err
	}

	// op, err := client.CreateBlockStorageVolume(ctx, v3.CreateBlockStorageVolumeRequest{
	// 	Name:                 c.Name,
	// 	Size:                 c.Size,
	// 	Labels:               c.Labels,
	// 	BlockStorageSnapshot: snapshot,
	// })
	// if err != nil {
	// 	return err
	// }
	// op, err = client.Wait(TODO, op, v3.OperationStateSuccess)
	// if err != nil {
	// 	return err
	// }

	// bs, err := client.GetBlockStorageVolume(TODO, op.Reference.ID)
	// if err != nil {
	// 	return err
	// }

	// if !globalstate.Quiet {
	// 	return (&blockstorageShowCmd{
	// 		cliCommandSettings: c.cliCommandSettings,
	// 		Name:               bs.ID.String(),
	// 	}).cmdRun(nil, nil)
	// }

	return nil
}

func init() {
	cobra.CheckErr(registerCLICommand(blockstorageSnapshotCmd, &blockstorageSnapshotCreateCmd{
		cliCommandSettings: defaultCLICmdSettings(),
	}))
}
