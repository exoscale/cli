package blockstorage

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	exocmd "github.com/exoscale/cli/cmd"
	"github.com/exoscale/cli/pkg/globalstate"
	"github.com/exoscale/cli/pkg/output"
	v3 "github.com/exoscale/egoscale/v3"
)

type blockStorageSnapshotShowOutput struct {
	CreatedAT time.Time                    `json:"created-at"`
	ID        v3.UUID                      `json:"id"`
	Volume    *v3.BlockStorageVolumeTarget `json:"volume"`
	Name      string                       `json:"name"`
	Size      string                       `json:"size"`
	State     v3.BlockStorageSnapshotState `json:"state"`
	Labels    map[string]string            `json:"labels"`
}

func (o *blockStorageSnapshotShowOutput) Type() string { return "Block Storage Volume Snapshot" }
func (o *blockStorageSnapshotShowOutput) ToJSON()      { output.JSON(o) }
func (o *blockStorageSnapshotShowOutput) ToText()      { output.Text(o) }
func (o *blockStorageSnapshotShowOutput) ToTable()     { output.Table(o) }

type blockStorageSnapshotShowCmd struct {
	exocmd.CliCommandSettings `cli-cmd:"-"`

	_ bool `cli-cmd:"show"`

	Name string      `cli-arg:"#" cli-usage:"NAME|ID"`
	Zone v3.ZoneName `cli-short:"z" cli-usage:"block storage zone"`
}

func (c *blockStorageSnapshotShowCmd) CmdAliases() []string { return exocmd.GShowAlias }

func (c *blockStorageSnapshotShowCmd) CmdShort() string { return "Show a Block Storage Volume details" }

func (c *blockStorageSnapshotShowCmd) CmdLong() string {
	return fmt.Sprintf(`This command shows a Block Storage Volume Snapshot details.

Supported output template annotations: %s`,
		strings.Join(output.TemplateAnnotations(&exocmd.InstanceShowOutput{}), ", "))
}

func (c *blockStorageSnapshotShowCmd) CmdPreRun(cmd *cobra.Command, args []string) error {
	exocmd.CmdSetZoneFlagFromDefault(cmd)
	return exocmd.CliCommandDefaultPreRun(c, cmd, args)
}

func (c *blockStorageSnapshotShowCmd) CmdRun(cmd *cobra.Command, _ []string) error {
	ctx := exocmd.GContext
	client, err := exocmd.SwitchClientZoneV3(ctx, globalstate.EgoscaleV3Client, c.Zone)
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

	return c.OutputFunc(&blockStorageSnapshotShowOutput{
		ID:        snapshot.ID,
		Name:      snapshot.Name,
		Size:      fmt.Sprintf("%d GiB", snapshot.Size),
		CreatedAT: snapshot.CreatedAT,
		State:     snapshot.State,
		Volume:    snapshot.BlockStorageVolume,
		Labels:    snapshot.Labels,
	}, nil)
}

func init() {
	cobra.CheckErr(exocmd.RegisterCLICommand(blockstorageSnapshotCmd, &blockStorageSnapshotShowCmd{
		CliCommandSettings: exocmd.DefaultCLICmdSettings(),
	}))
}
