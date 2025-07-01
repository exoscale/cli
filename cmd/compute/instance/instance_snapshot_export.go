package instance

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

type instanceSnapshotExportOutput struct {
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
}

func (o *instanceSnapshotExportOutput) ToJSON()  { output.JSON(o) }
func (o *instanceSnapshotExportOutput) ToText()  { output.Text(o) }
func (o *instanceSnapshotExportOutput) ToTable() { output.Table(o) }

type instanceSnapshotExportCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"export"`

	ID string `cli-arg:"#"`

	Zone string `cli-short:"z" cli-usage:"snapshot zone"`
}

func (c *instanceSnapshotExportCmd) CmdAliases() []string { return nil }

func (c *instanceSnapshotExportCmd) CmdShort() string {
	return "Export a Compute instance snapshot"
}

func (c *instanceSnapshotExportCmd) CmdLong() string {
	return fmt.Sprintf(`This command exports a Compute instance snapshot.
	
Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&instanceSnapshotExportOutput{}), ", "))
}

func (c *instanceSnapshotExportCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *instanceSnapshotExportCmd) CmdRun(_ *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, v3.ZoneName(c.Zone))
	if err != nil {
		return err
	}

	snapshots, err := client.ListSnapshots(ctx)
	if err != nil {
		return err
	}
	snapshot, err := snapshots.FindSnapshot(c.ID)
	if err != nil {
		return err
	}

	op, err := client.ExportSnapshot(ctx, snapshot.ID)
	utils.DecorateAsyncOperation(fmt.Sprintf("Exporting snapshot %s...", c.ID), func() {
		_, err = client.Wait(ctx, op, v3.OperationStateSuccess)
	})
	if err != nil {
		return err
	}

	exportedSnapshot, err := client.GetSnapshot(ctx, snapshot.ID)
	if err != nil {
		return err
	}

	if !globalstate.Quiet {
		return c.OutputFunc(
			&instanceSnapshotExportOutput{
				URL:      exportedSnapshot.Export.PresignedURL,
				Checksum: exportedSnapshot.Export.Md5sum,
			},
			nil)
	}

	return nil
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(instanceSnapshotCmd, &instanceSnapshotExportCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
